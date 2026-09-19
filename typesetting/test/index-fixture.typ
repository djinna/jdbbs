// Fixture for the back-of-book index piece (series-template.typ, 5.13).
// Compile from the repo root with both binaries:
//   typst compile --root . --font-path typesetting/fonts typesetting/test/index-fixture.typ /tmp/idx.pdf
// Expected: body pages 1–4, then an "Index" recto with letter heads, run-in
// subentries, a range (1–3), a see-reference and a see-also.
#import "/typesetting/templates/series-template.typ": *
#let config = merge-config((index: true))

#show: book.with(config: config, title: "Index Fixture", author: "Test Author")

= Chapter One
#first-para[
Rice was the region's dietary staple.#index[Rice] The khlongs carried water
to the paddies.#index("Khlongs", "as irrigation") An artificial ghost, or
#emph[beaing], haunts the machine.#index("Beaings", see-also: "ghosts, artificial")
Typesetting is done in Typst.#index("Typesetting", see: "Typst", locator: false)#index[Typst]
The Éminence grise sorts under E.#index[Éminence grise] A leading article is
ignored.#index[The Factory Pass] Numbers go under the shared head.#index[3D printing] A code-like heading wraps.#index[AmaStore_L47_HeartVariant1.0_ExtendedEditionBuild/release-candidate]
]
#block[
Second paragraph, same page, mentions rice again.#index[Rice]
]

#pagebreak()
Page two mentions rice#index[Rice] and the khlongs#index("Khlongs", "as irrigation")
and khlongs as commons.#index("Khlongs", "as commons")

#pagebreak()
Page three: rice#index[Rice] for the range, and a ghost.#index("ghosts, artificial")

#pagebreak()
Page four: only Typst.#index[Typst]
