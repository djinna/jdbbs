package srv

import (
	"encoding/json"
	"fmt"
	"html"
	"log/slog"
	"strings"

	"srv.exe.dev/db/dbgen"
)

const mailKindStoreAddon = "store_addon"

// storeItemsSummary renders an order's items JSON as "+3 builds ×2, +6 months storage".
func storeItemsSummary(itemsJSON string) string {
	var items []struct {
		LookupKey string `json:"lookup_key"`
		Quantity  int64  `json:"quantity"`
	}
	_ = json.Unmarshal([]byte(itemsJSON), &items)
	var parts []string
	for _, it := range items {
		def := storeItemByKey(it.LookupKey)
		name := it.LookupKey
		if def != nil {
			name = def.Name
		}
		if it.Quantity > 1 {
			name += fmt.Sprintf(" ×%d", it.Quantity)
		}
		parts = append(parts, name)
	}
	return strings.Join(parts, ", ")
}

func fmtUSD(cents int64) string {
	if cents%100 == 0 {
		return fmt.Sprintf("$%d", cents/100)
	}
	return fmt.Sprintf("$%d.%02d", cents/100, cents%100)
}

// sendAddonEmail confirms an add-on purchase against an existing pass:
// what was bought, the new credit count and expiry. Fire-and-forget.
func (s *Server) sendAddonEmail(pass dbgen.Pass, order dbgen.StoreOrder, title, portalURL string) {
	if s.Email == nil {
		slog.Warn("store: add-on fulfilled but email not configured", "pass_id", pass.ID)
		return
	}
	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("add-on email panic", "recover", rec)
			}
		}()
		if title == "" {
			title = "your book"
		}
		bought := storeItemsSummary(order.Items)
		credits := passCreditsRemaining(pass)
		expires := pass.ExpiresAt.UTC().Format("2 January 2006")
		subject := fmt.Sprintf("Added to your Factory Pass: %s", bought)

		var t strings.Builder
		fmt.Fprintf(&t, "Hi %s,\n\n", firstName(pass.CustomerName))
		fmt.Fprintf(&t, "Thanks — %s has been added to your Factory Pass for %s.\n\n", bought, title)
		fmt.Fprintf(&t, "Builds remaining:  %d\n", credits)
		fmt.Fprintf(&t, "Pass live until:   %s\n", expires)
		fmt.Fprintf(&t, "Paid:              %s\n", fmtUSD(order.AmountTotal))
		fmt.Fprintf(&t, "Your factory:      %s\n\n", portalURL)
		t.WriteString("Stripe sends the card receipt separately.\n\n")
		t.WriteString(emailSignoffText())

		var h strings.Builder
		h.WriteString(emailP(fmt.Sprintf("Hi %s,", html.EscapeString(firstName(pass.CustomerName)))))
		h.WriteString(emailP(fmt.Sprintf("Thanks &mdash; <b>%s</b> has been added to your Factory Pass for <b>%s</b>.", html.EscapeString(bought), html.EscapeString(title))))
		h.WriteString(emailKV([][2]string{
			{"Builds remaining", fmt.Sprint(credits)},
			{"Pass live until", expires},
			{"Paid", fmtUSD(order.AmountTotal)},
		}))
		h.WriteString(emailButton(portalURL, "Open your factory"))
		h.WriteString(emailSmall("Stripe sends the card receipt separately."))
		h.WriteString(emailSignoff())
		htmlBody := emailShell(h.String(), emailShellOpts{Kicker: "Factory Pass · add-on", Title: title})

		if err := s.mail(mailMeta{Kind: mailKindStoreAddon, RefType: "pass", RefID: mailRef(pass.ID), TriggeredBy: "stripe"},
			[]string{pass.CustomerEmail}, nil, subject, t.String(), htmlBody); err != nil {
			slog.Error("add-on email failed", "err", err, "pass_id", pass.ID)
		}
	}()
}
