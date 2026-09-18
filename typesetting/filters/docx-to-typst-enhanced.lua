-- docx-to-typst-enhanced.lua
-- Enhanced Pandoc Lua filter with edge case handling
-- 
-- Usage: pandoc input.docx --lua-filter=docx-to-typst-enhanced.lua -o output.typ

-- Helper: wrap content in a Typst function call
local function typst_call(func, content)
  return pandoc.RawInline('typst', '#' .. func .. '[') 
    .. content 
    .. pandoc.RawInline('typst', ']')
end

local function typst_block(func, content)
  return pandoc.RawBlock('typst', '#' .. func .. '[' .. pandoc.utils.stringify(content) .. ']')
end

-- =============================================================================
-- CHARACTER (INLINE) STYLE MAPPINGS
-- =============================================================================

-- Declared custom styles from the book spec, passed by the server via
-- --metadata-file as custom_styles: [{word_style, ident, type}]. Keyed by
-- normalized Word style name → Typst function name. Consulted after the
-- hardcoded maps so a spec-declared "Field Note" becomes #field-note[...].
local declared_para_styles = {}
local declared_char_styles = {}

-- Normalize style name for matching
local function normalize_style(name)
  if not name then return "" end
  return name:lower():gsub("[%s%-_]", "")
end

local function normalize_text_for_lookup(text)
  if not text then return "" end
  -- normalize NBSP to regular space first
  local normalized = text:gsub("\194\160", " "):gsub("\160", " ")
  return normalized:lower():gsub("%s+", " "):gsub("^%s+", ""):gsub("%s+$", "")
end

local function list_body_key(text)
  local key = normalize_text_for_lookup(text)
  -- remove manual list prefixes (1. / 1) / a) / a. / - / * / •)
  key = key:gsub("^%d+[%.%)%s]+", "")
  key = key:gsub("^[a-zA-Z][%.%)%s]+", "")
  key = key:gsub("^[%*%-•∙◦▪▫◾◽]+%s+", "")
  return key
end

local function list_body_text(text)
  local normalized = normalize_text_for_lookup(text)
  normalized = normalized:gsub("^%d+[%.%)%s]+", "")
  normalized = normalized:gsub("^[a-zA-Z][%.%)%s]+", "")
  normalized = normalized:gsub("^[%*%-•∙◦▪▫◾◽]+%s+", "")
  return normalized
end

local edge_decisions = {}
local edge_decisions_loaded = false

local function decision_bucket(decision_type)
  local t = decision_type or "manual_list"
  if not edge_decisions[t] then
    edge_decisions[t] = {}
  end
  return edge_decisions[t]
end

local function add_decision(decision_type, text, decision)
  if not text or not decision then return end
  local normalized_type = tostring(decision_type or "manual_list"):lower()
  local normalized_decision = tostring(decision):lower()
  local bucket = decision_bucket(normalized_type)

  local key_full = normalize_text_for_lookup(text)
  if key_full ~= "" then
    bucket[key_full] = normalized_decision
  end

  if normalized_type == "manual_list" then
    local key_body = list_body_key(text)
    if key_body ~= "" then
      bucket[key_body] = normalized_decision
    end
  end
end

local function get_decision(decision_type, text)
  local normalized_type = tostring(decision_type or "manual_list"):lower()
  local bucket = edge_decisions[normalized_type]
  if not bucket then return nil end

  local key_full = normalize_text_for_lookup(text)
  if bucket[key_full] then
    return bucket[key_full]
  end

  if normalized_type == "manual_list" then
    local key_body = list_body_key(text)
    if bucket[key_body] then
      return bucket[key_body]
    end
  end

  return nil
end

local function apply_strip_substrings(text, decision_type)
  local bucket = edge_decisions[tostring(decision_type):lower()]
  if not bucket or not text or text == "" then
    return text, false
  end

  local out = text
  local changed = false

  for key, decision in pairs(bucket) do
    if decision == "strip" and key and key ~= "" then
      local lowered = out:lower()
      local s = lowered:find(key, 1, true)
      while s do
        local e = s + #key - 1
        out = out:sub(1, s - 1) .. out:sub(e + 1)
        changed = true
        lowered = out:lower()
        s = lowered:find(key, 1, true)
      end
    end
  end

  if changed then
    out = out:gsub("%s+", " "):gsub("^%s+", ""):gsub("%s+$", "")
  end

  return out, changed
end

local function load_edge_decisions_map(doc)
  local meta_root = doc.meta or doc
  local meta = meta_root["edge-decisions-map-file"] or meta_root["edge_decisions_map_file"]
  if not meta then return 0 end

  local path = pandoc.utils.stringify(meta)
  if not path or path == "" then return 0 end

  local file = io.open(path, "r")
  if not file then
    io.stderr:write("[docx-to-typst-enhanced] warning: could not open edge decisions map file: " .. path .. "\n")
    return 0
  end

  local count = 0
  for line in file:lines() do
    local decision_type, decision, text = line:match("^([^\t]+)\t([^\t]+)\t(.+)$")
    if decision_type and decision and text then
      add_decision(decision_type, text, decision)
      count = count + 1
    end
  end
  file:close()

  return count
end

local function load_edge_decisions(doc)
  local meta_root = doc.meta or doc
  local meta = meta_root["edge-decisions-file"] or meta_root["edge_decisions_file"]
  if not meta then return end

  local path = pandoc.utils.stringify(meta)
  if not path or path == "" then return end

  local file = io.open(path, "r")
  if not file then
    io.stderr:write("[docx-to-typst-enhanced] warning: could not open edge decisions file: " .. path .. "\n")
    return
  end

  local raw = file:read("*a")
  file:close()

  local ok, parsed = pcall(pandoc.json.decode, raw)
  if not ok or type(parsed) ~= "table" then
    io.stderr:write("[docx-to-typst-enhanced] warning: could not parse edge decisions JSON\n")
    return
  end

  local decisions = parsed.decisions or parsed
  if type(decisions) ~= "table" then
    return
  end

  local supported_types = {
    manual_list = true,
    colored_text = true,
    highlighted_text = true,
  }

  local count = 0
  for _, item in ipairs(decisions) do
    if type(item) == "table" and item.type and supported_types[item.type] and item.text and item.decision then
      add_decision(item.type, item.text, item.decision)
      count = count + 1
    end
  end

  if count > 0 then
    edge_decisions_loaded = true
  end
end

-- Map of Word character styles to Typst functions
local char_style_map = {
  -- Small caps / acronyms
  ["smallcaps"] = "sc",
  ["acronym"] = "sc",
  ["stonehand"] = "sc",  -- InDesign style name
  
  -- Emphasis variations
  ["emphasis"] = nil,  -- Let Pandoc handle as native italic
  ["italic"] = nil,
  ["ital"] = nil,
  ["strong"] = nil,    -- Let Pandoc handle as native bold
  ["bold"] = nil,
  
  -- Explicit styled
  ["bolditalic"] = "bold-ital",
  ["boldital"] = "bold-ital",
  
  -- Titles
  ["booktitle"] = "booktitle",
  ["worktitle"] = "booktitle",
  ["articletitle"] = "articletitle",
  
  -- Foreign text
  ["foreign"] = "foreign",
  ["foreignword"] = "foreign",
  
  -- Code
  ["code"] = nil,  -- Let Pandoc handle as Code
  ["codechar"] = nil,
  ["monospace"] = nil,
  
  -- Special
  ["tracked"] = "tracked",
  ["letterspaced"] = "tracked",
  ["allcaps"] = "allcaps",

  -- Project custom inline styles
  ["metadatac"] = "metadata-c",
}

function Span(el)
  -- Check for custom-style attribute (how Pandoc represents Word styles)
  local style = el.attributes["custom-style"]
  if not style then
    -- Check for color or highlighting
    if el.attributes.color or el.attributes.highlight then
      local span_text = pandoc.utils.stringify(el.content)
      local decision_type = el.attributes.color and "colored_text" or "highlighted_text"
      local decision = get_decision(decision_type, span_text)

      if decision == "strip" then
        return {}
      elseif decision == "keep" then
        return el
      elseif decision == "convert" then
        return {
          pandoc.RawInline('typst', '// Editorial decision applied: convert ' .. decision_type .. ' '),
          el,
        }
      end

      -- Default: add editorial comment for unresolved formatting
      local comment = ""
      if el.attributes.color then
        comment = comment .. "color:" .. el.attributes.color .. " "
      end
      if el.attributes.highlight then
        comment = comment .. "highlight:" .. el.attributes.highlight
      end
      return {
        pandoc.RawInline('typst', '#text(fill: red)['),
        el,
        pandoc.RawInline('typst', '] // Editorial: ' .. comment)
      }
    end
    return nil
  end
  
  local normalized = normalize_style(style)
  local typst_func = char_style_map[normalized] or declared_char_styles[normalized]
  
  if typst_func then
    -- Wrap in Typst function while preserving original inline nodes so
    -- the Typst writer can escape/render content safely.
    local out = { pandoc.RawInline('typst', '#' .. typst_func .. '[') }
    for _, item in ipairs(el.content) do
      table.insert(out, item)
    end
    table.insert(out, pandoc.RawInline('typst', ']'))
    return out
  end
  
  -- Return unchanged if no mapping
  return nil
end

-- =============================================================================
-- PARAGRAPH STYLE MAPPINGS
-- =============================================================================

local para_style_map = {
  -- Section break
  ["sectionbreak"] = "section-break",
  ["break"] = "section-break",
  ["scenebreak"] = "section-break",
  
  -- First paragraph (no indent)
  ["firstparagraph"] = "first-para",
  ["basicpara1st"] = "first-para",
  ["noindent"] = "first-para",
  
  -- Epigraph
  ["epigraph"] = "epigraph",
  ["epi"] = "epigraph",
  
  -- Poem/verse
  ["poem"] = "poem",
  ["verse"] = "poem",
  ["poetry"] = "poem",

  -- Remaining house styles shipped in the generated Word template
  ["codeblock"] = "code-block",
  ["copyright"] = "copyright-page",

  -- Project custom paragraph styles
  ["tweetp"] = "tweet-p",
  ["metadatap"] = "metadata-p",

  -- ASCII-art / preformatted tweet blocks
  ["tweetpascii"] = "tweet-p-ascii",
}

function Div(el)
  local style = el.attributes["custom-style"]
  if not style then return nil end
  
  local normalized = normalize_style(style)
  local typst_func = para_style_map[normalized] or declared_para_styles[normalized]
  
  if typst_func == "section-break" then
    -- Section break is just a marker, no content
    return pandoc.RawBlock('typst', '\n#section-break\n')
  elseif typst_func then
    -- Wrap block content while preserving original block nodes so the Typst
    -- writer handles escaping and inline syntax correctly.
    local out = { pandoc.RawBlock('typst', '\n#' .. typst_func .. '[') }
    for _, item in ipairs(el.content) do
      table.insert(out, item)
    end
    table.insert(out, pandoc.RawBlock('typst', ']\n'))
    return out
  end
  
  return nil
end

-- =============================================================================
-- EDGE CASE HANDLING
-- =============================================================================

-- Detect manual section breaks in paragraphs
function Para(el)
  local raw_text = pandoc.utils.stringify(el.content)
  local strip_applied_text, strip_applied = apply_strip_substrings(raw_text, "colored_text")
  if strip_applied then
    local parsed = pandoc.read(strip_applied_text, "markdown")
    if parsed.blocks and #parsed.blocks > 0 and parsed.blocks[1].t == "Para" then
      return parsed.blocks[1]
    end
    return pandoc.Para({ pandoc.Str(strip_applied_text) })
  end

  local text = raw_text:gsub("%s+", " ")
  text = text:gsub("^%s+", ""):gsub("%s+$", "") -- trim

  -- Common section break patterns
  if text:match("^[*%-–—]+%s*[*%-–—]+%s*[*%-–—]+$") or
     text == "* * *" or
     text == "***" or
     text == "---" or
     text == "———" or
     (text:match("^[%*%s]+$") and #text > 2) then
    return pandoc.RawBlock('typst', '\n#section-break\n')
  end

  -- Single centered ornament characters
  if (text == "❦" or text == "❧" or text == "⁂" or text == "§") and
     el.content[1] and el.content[1].t == "Str" then
    return pandoc.RawBlock('typst', '\n#section-break\n')
  end

  -- Manual list items: apply optional editorial decisions (hook), else flag for review
  local manual_numbered = text:match("^%d+[%.%)%s]+.+")
  local manual_lettered = text:match("^[a-zA-Z][%.%)%s]+.+")
  local manual_bullet = text:match("^[%*%-•∙◦▪▫◾◽]+%s+.+")
  if manual_numbered or manual_lettered or manual_bullet then
    local decision = get_decision("manual_list", text)

    if decision == "strip" then
      return {}
    elseif decision == "keep" then
      return nil
    elseif decision == "convert" then
      local body = list_body_text(text)
      if body == "" then
        body = normalize_text_for_lookup(text)
      end

      if manual_numbered or manual_lettered then
        return {
          pandoc.RawBlock('typst', '// Editorial decision applied: convert manual numbered/lettered list item'),
          pandoc.RawBlock('typst', '+ ' .. body),
        }
      end

      return {
        pandoc.RawBlock('typst', '// Editorial decision applied: convert manual bullet list item'),
        pandoc.RawBlock('typst', '- ' .. body),
      }
    end

    return {
      pandoc.RawBlock('typst', '// Editorial review: manual list item detected; keep/strip/convert decision required'),
      el,
    }
  end

  return nil
end

-- =============================================================================
-- SPECIAL HANDLING
-- =============================================================================

-- Convert horizontal rules to section breaks
function HorizontalRule()
  return pandoc.RawBlock('typst', '\n#section-break\n')
end

-- Handle code blocks
function CodeBlock(el)
  -- Already handled well by Pandoc's native Typst writer
  return nil
end

-- Handle block quotes  
function BlockQuote(el)
  local content = pandoc.utils.stringify(el.content)
  return pandoc.RawBlock('typst', '\n#blockquote[\n' .. content .. '\n]\n')
end

-- =============================================================================
-- DOCUMENT WRAPPER
-- =============================================================================

-- TRK-DEV-012 Phase B: per-chapter `#set-story-info()` injection for the
-- Typst output. The Go pipeline writes a chapters.json metadata file and
-- passes it via --metadata-file when the book spec has spec.chapters[].
-- We read it here and emit a RawBlock before every level-1 heading, indexed
-- by source-order h1 count. Single-author books (no metadata) are unchanged.
local chapter_meta = nil
local h1_count = 0

local function typst_escape(s)
  if not s then return "" end
  s = s:gsub('\\', '\\\\')
  s = s:gsub('"', '\\"')
  return s
end

local function load_declared_styles(meta)
  if not meta.custom_styles then return end
  for _, cs in ipairs(meta.custom_styles) do
    local word = cs.word_style and pandoc.utils.stringify(cs.word_style) or ""
    local ident = cs.ident and pandoc.utils.stringify(cs.ident) or ""
    local kind = cs.type and pandoc.utils.stringify(cs.type) or "paragraph"
    if word ~= "" and ident ~= "" then
      if kind == "character" then
        declared_char_styles[normalize_style(word)] = ident
      else
        declared_para_styles[normalize_style(word)] = ident
      end
    end
  end
end

-- P4 front matter (docs/reviews/P4-FRONT-MATTER-PLAN-2026-09-18.md). The Go
-- pipeline computes the book map (srv/bookmap.go) and passes it as metadata:
--   book_map = { toc = bool,
--                sections = { {title=, kind=front|body|back|title|toc}, ... },  -- every H1 in order
--                untitled_front = { {name=, paras=, drop=}, ... } }             -- blocks before the first H1
-- We turn it into template calls: #front-piece(...)[...] around the untitled
-- pieces, #contents-page(), #front-section() / #start-body() / #start-back()
-- before the right H1s, and drop title/toc sections outright.
local book_map = nil

local function meta_bool(v)
  if v == nil then return nil end
  if type(v) == "boolean" then return v end
  local s = pandoc.utils.stringify(v)
  return s == "true"
end

local function load_book_map(meta)
  if not meta.book_map then return end
  local bm = { sections = {}, untitled = {}, toc = false }
  local t = meta_bool(meta.book_map.toc)
  bm.toc = t == true
  bm.parts = meta_bool(meta.book_map.parts) == true
  if meta.book_map.sections then
    for _, sec in ipairs(meta.book_map.sections) do
      table.insert(bm.sections, {
        title = sec.title and pandoc.utils.stringify(sec.title) or "",
        kind = sec.kind and pandoc.utils.stringify(sec.kind) or "body",
      })
    end
  end
  if meta.book_map.untitled_front then
    for _, u in ipairs(meta.book_map.untitled_front) do
      table.insert(bm.untitled, {
        name = u.name and pandoc.utils.stringify(u.name) or "dedication",
        paras = tonumber(u.paras and pandoc.utils.stringify(u.paras) or "1") or 1,
        drop = meta_bool(u.drop) == true,
      })
    end
  end
  book_map = bm
end

-- How many manuscript paragraphs a block stands for (lists collapse several
-- w:p into one block; RawBlocks are our own wrappers and count for nothing).
local function block_para_count(b)
  if b.t == "RawBlock" then return 0 end
  if b.t == "BulletList" or b.t == "OrderedList" then
    return math.max(1, #b.content)
  end
  return 1
end

-- Our Div/Para handlers wrap styled paragraphs as RawBlock("#func[") … RawBlock("]").
-- A piece boundary must never fall inside such a wrapper.
local function raw_depth_delta(b)
  if b.t ~= "RawBlock" then return 0 end
  local t = b.text
  if t:match("^%s*#[%w%-]+%[%s*$") then return 1 end
  if t:match("^%s*%]%s*$") then return -1 end
  return 0
end

-- Consume blocks[i..] for one piece of `paras` paragraphs; returns the next
-- index. `take` receives each consumed block (nil to drop).
local function consume_piece(blocks, i, pre_end, paras, last_piece, take)
  local remaining, depth = paras, 0
  while i <= pre_end and (remaining > 0 or depth > 0 or last_piece) do
    local b = blocks[i]
    depth = depth + raw_depth_delta(b)
    remaining = remaining - block_para_count(b)
    if take then take(b) end
    i = i + 1
  end
  return i
end

local function raw(s) return pandoc.RawBlock("typst", s) end

-- Restructure doc.blocks per the book map. Returns the new block list.
local function apply_book_map(blocks)
  if not book_map then return blocks end
  local out = {}
  local first_h1 = nil
  for i, b in ipairs(blocks) do
    if b.t == "Header" and b.level == 1 then first_h1 = i; break end
  end
  if not first_h1 then
    -- No section headings at all: the whole file is body text.
    table.insert(out, raw('#start-body()'))
    for _, b in ipairs(blocks) do table.insert(out, b) end
    return out
  end
  local pre_end = first_h1 - 1

  -- Untitled front matter: group the leading blocks into the map's pieces.
  -- Only real paragraphs count; our own RawBlocks (e.g. #set-story-info
  -- injected before the first H1) are passed through untouched.
  local pre_paras = 0
  for k = 1, pre_end do pre_paras = pre_paras + block_para_count(blocks[k]) end
  local i = 1
  local pending = {} -- RawBlocks held back to stay glued to the next heading
  if pre_paras == 0 then
    while i <= pre_end do table.insert(pending, blocks[i]); i = i + 1 end
  elseif pre_end >= 1 then
    if #book_map.untitled == 0 then
      -- Nothing declared: keep the blocks as one untitled piece.
      table.insert(out, raw('#front-piece(kind: "untitled")['))
      while i <= pre_end do table.insert(out, blocks[i]); i = i + 1 end
      table.insert(out, raw(']'))
    else
      for pi, piece in ipairs(book_map.untitled) do
        if i > pre_end then break end
        local last_piece = (pi == #book_map.untitled)
        if piece.drop then
          -- Byline / typed copyright: generated from the transmittal instead.
          i = consume_piece(blocks, i, pre_end, piece.paras, false, nil)
        else
          table.insert(out, raw(string.format('#front-piece(kind: "%s")[', typst_escape(piece.name))))
          i = consume_piece(blocks, i, pre_end, piece.paras, last_piece, function(b) table.insert(out, b) end)
          table.insert(out, raw(']'))
        end
      end
    end
  end

  -- Contents page after the untitled pieces, before the first kept heading.
  local toc_pending = book_map.toc

  -- H1 sections. RawBlocks directly before a heading (e.g. #set-story-info
  -- from the anthology pass) are held back so our hooks land before them and
  -- they stay glued to their heading.
  local h1_index = 0
  local dropping = false
  local body_started, back_started = false, false
  local function flush_pending()
    for _, r in ipairs(pending) do table.insert(out, r) end
    pending = {}
  end
  while i <= #blocks do
    local b = blocks[i]
    if b.t == "Header" and b.level == 1 then
      h1_index = h1_index + 1
      local sec = book_map.sections[h1_index]
      local kind = sec and sec.kind or "body"
      dropping = (kind == "title" or kind == "toc")
      if dropping then
        pending = {}
      else
        if toc_pending then
          table.insert(out, raw('#contents-page()'))
          toc_pending = false
        end
        if kind == "front" then
          table.insert(out, raw('#front-section()'))
        elseif kind == "back" then
          if not body_started then table.insert(out, raw('#start-body()')); body_started = true end
          if not back_started then table.insert(out, raw('#start-back()')); back_started = true end
        else
          if not body_started then table.insert(out, raw('#start-body()')); body_started = true end
        end
        flush_pending()
        -- Parts books: H1 is the part opener, so a front/back section
        -- (Foreword, Notes) is demoted one level to read as a chapter.
        if book_map.parts and kind ~= "body" then b.level = 2 end
        table.insert(out, b)
      end
    elseif not dropping then
      if b.t == "RawBlock" then
        table.insert(pending, b)
      else
        flush_pending()
        table.insert(out, b)
      end
    end
    i = i + 1
  end
  flush_pending()
  if toc_pending then table.insert(out, raw('#contents-page()')) end
  if not body_started then table.insert(out, raw('#start-body()')) end
  return out
end

function Meta(meta)
  load_declared_styles(meta)
  load_book_map(meta)
  local map_count = load_edge_decisions_map(meta)
  if map_count > 0 then
    edge_decisions_loaded = true
  else
    load_edge_decisions(meta)
  end

  if meta.chapters then
    chapter_meta = {}
    for _, c in ipairs(meta.chapters) do
      local t = ""
      local a = ""
      if c.title then t = pandoc.utils.stringify(c.title) end
      if c.author then a = pandoc.utils.stringify(c.author) end
      table.insert(chapter_meta, { title = t, author = a })
    end
  end
  return nil
end

function Header(el)
  if el.level ~= 1 then return nil end
  if not chapter_meta then return nil end
  h1_count = h1_count + 1
  local c = chapter_meta[h1_count]
  if not c then return nil end
  local title = typst_escape(c.title)
  local author = typst_escape(c.author)
  local raw = string.format('#set-story-info(title: "%s", author: "%s")\n', title, author)
  return { pandoc.RawBlock("typst", raw), el }
end

local function is_convert_comment_block(block)
  if block.t ~= "RawBlock" or block.format ~= "typst" then return false end
  return block.text:match("^%s*//%s*Editorial decision applied: convert manual") ~= nil
end

local function extract_typst_list_item(block)
  if block.t ~= "RawBlock" or block.format ~= "typst" then return nil, nil end
  local marker, body = block.text:match("^%s*([%+%-])%s+(.+)%s*$")
  if not marker then return nil, nil end
  return marker, body
end

local function normalize_converted_list_blocks(blocks)
  local out = {}
  local current = nil
  local current_marker = nil
  local pending_convert_comment = false

  local function flush_current()
    if not current or #current == 0 then
      current = nil
      current_marker = nil
      return
    end

    table.insert(out, pandoc.RawBlock('typst', '// Editorial decision applied: convert manual list block'))
    table.insert(out, pandoc.RawBlock('typst', table.concat(current, "\n")))
    current = nil
    current_marker = nil
  end

  for _, block in ipairs(blocks) do
    if is_convert_comment_block(block) then
      pending_convert_comment = true
    else
      local marker, body = extract_typst_list_item(block)
      if marker and (pending_convert_comment or current) then
        local line = marker .. " " .. body
        if current and marker == current_marker then
          table.insert(current, line)
        else
          flush_current()
          current = { line }
          current_marker = marker
        end
        pending_convert_comment = false
      else
        flush_current()
        if pending_convert_comment then
          table.insert(out, pandoc.RawBlock('typst', '// Editorial decision applied: convert manual list item'))
          pending_convert_comment = false
        end
        table.insert(out, block)
      end
    end
  end

  flush_current()

  if pending_convert_comment then
    table.insert(out, pandoc.RawBlock('typst', '// Editorial decision applied: convert manual list item'))
  end

  return out
end

-- Add template import at document start
function Pandoc(doc)
  doc.blocks = normalize_converted_list_blocks(doc.blocks)
  doc.blocks = apply_book_map(doc.blocks)

  local header = pandoc.RawBlock('typst', [[
#import "/templates/series-template.typ": *

#show: book.with(
  title: "TITLE",
  author: "AUTHOR",
)

]])

  table.insert(doc.blocks, 1, header)

  if edge_decisions_loaded then
    table.insert(doc.blocks, 2, pandoc.RawBlock('typst', '// Editorial decisions loaded: manual list/color/highlight keep/strip/convert hooks active'))
  end

  return doc
end

-- =============================================================================
-- RUN FILTERS IN ORDER
-- =============================================================================

-- Pandoc runs filters in this order
return {
  { Meta = Meta },
  { Span = Span },
  { Div = Div },
  { Para = Para },
  { Header = Header },
  { HorizontalRule = HorizontalRule },
  { CodeBlock = CodeBlock },
  { BlockQuote = BlockQuote },
  { Pandoc = Pandoc }
}