-- docx-to-typst.lua
-- Pandoc Lua filter to map Word styles to Typst functions
-- 
-- Usage: pandoc input.docx --lua-filter=docx-to-typst.lua -o output.typ

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

-- Normalize style name for matching
local function normalize_style(name)
  if not name then return "" end
  return name:lower():gsub("[%s%-_]", "")
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
}

function Span(el)
  -- Check for custom-style attribute (how Pandoc represents Word styles)
  local style = el.attributes["custom-style"]
  if not style then return nil end
  
  local normalized = normalize_style(style)
  local typst_func = char_style_map[normalized]
  
  if typst_func then
    -- Wrap in Typst function
    local content = pandoc.utils.stringify(el.content)
    return pandoc.RawInline('typst', '#' .. typst_func .. '[' .. content .. ']')
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
}

function Div(el)
  local style = el.attributes["custom-style"]
  if not style then return nil end
  
  local normalized = normalize_style(style)
  local typst_func = para_style_map[normalized]
  
  if typst_func == "section-break" then
    -- Section break is just a marker, no content
    return pandoc.RawBlock('typst', '\n#section-break\n')
  elseif typst_func then
    -- Wrap paragraph content
    local content = pandoc.utils.stringify(el.content)
    return pandoc.RawBlock('typst', '\n#' .. typst_func .. '[\n' .. content .. '\n]\n')
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

-- Add template import at document start
function Pandoc(doc)
  local header = pandoc.RawBlock('typst', [[
#import "templates/series-template.typ": *

#show: book.with(
  title: "TITLE",
  author: "AUTHOR",
)

]])
  
  table.insert(doc.blocks, 1, header)
  return doc
end

-- =============================================================================
-- METADATA
-- =============================================================================

-- Could extract title/author from Word metadata
-- For now, uses placeholders that user fills in
