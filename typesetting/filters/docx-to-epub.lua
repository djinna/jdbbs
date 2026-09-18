-- docx-to-epub.lua — P4 book map for the EPUB build.
-- (docs/reviews/P4-FRONT-MATTER-PLAN-2026-09-18.md, step 4)
--
-- The Go pipeline passes the same `book_map` metadata the typst filter gets:
--   book_map = { toc = bool,
--                sections = { {title=, kind=front|body|back|title|toc}, ... },
--                untitled_front = { {name=, paras=, drop=}, ... } }
-- Here we:
--   * drop the byline / typed copyright / typed "Contents" / repeated title
--     section (the EPUB has its own title page and nav);
--   * give each untitled front piece (dedication, epigraph) its own section
--     with an epub:type, so it lands in the landmarks and the reading order;
--   * tag every H1 with epub:type (foreword, chapter, appendix, …) so pandoc
--     emits frontmatter / bodymatter / backmatter sections and a landmarks nav.

local book_map = nil

local function meta_bool(v)
  if v == nil then return nil end
  if type(v) == "boolean" then return v end
  return pandoc.utils.stringify(v) == "true"
end

function Meta(meta)
  if not meta.book_map then return end
  local bm = { sections = {}, untitled = {} }
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

-- EPUB structural semantics vocabulary, from the heading text.
local function norm(s)
  s = s:lower():gsub("[%p]", ""):gsub("%s+", " ")
  return s:match("^%s*(.-)%s*$")
end

local EXACT = {
  foreword = "foreword", preface = "preface", prologue = "prologue",
  introduction = "introduction", acknowledgments = "acknowledgments",
  acknowledgements = "acknowledgments", dedication = "dedication",
  epigraph = "epigraph", epilogue = "epilogue", afterword = "afterword",
  conclusion = "conclusion", appendix = "appendix", appendices = "appendix",
  notes = "endnotes", endnotes = "endnotes", bibliography = "bibliography",
  references = "bibliography", ["works cited"] = "bibliography",
  ["further reading"] = "bibliography", glossary = "glossary", index = "index",
  colophon = "colophon", contributors = "contributors",
  ["about the author"] = "contributors", ["about the authors"] = "contributors",
  ["list of illustrations"] = "loi", illustrations = "loi",
  ["list of tables"] = "lot", ["list of figures"] = "loi",
  credits = "credits", permissions = "credits", chronology = "chronology",
}

local function epub_type(title, kind)
  local n = norm(title)
  if EXACT[n] then return EXACT[n] end
  if n:match("^appendix") then return "appendix" end
  if n:match("^list of ") then return "loi" end
  if n:match("^part ") or n:match("^book ") then return "part" end
  if kind == "front" then return "preface" end
  if kind == "back" then return "afterword" end
  return "chapter"
end

local function block_para_count(b)
  if b.t == "BulletList" or b.t == "OrderedList" then return math.max(1, #b.content) end
  return 1
end

local function titlecase(s)
  return (s:gsub("^%l", string.upper))
end

function Pandoc(doc)
  if not book_map then return doc end
  local blocks = doc.blocks
  local out = {}
  local first_h1 = nil
  for i, b in ipairs(blocks) do
    if b.t == "Header" and b.level == 1 then first_h1 = i; break end
  end
  local pre_end = first_h1 and (first_h1 - 1) or #blocks

  -- Untitled front pieces → own sections with a hidden heading. The heading
  -- gives the piece a nav entry ("Dedication") and an epub:type; CSS hides it.
  local i = 1
  if pre_end >= 1 then
    if #book_map.untitled == 0 then
      while i <= pre_end do table.insert(out, blocks[i]); i = i + 1 end
    else
      for pi, piece in ipairs(book_map.untitled) do
        local last = (pi == #book_map.untitled)
        local body = {}
        local remaining = piece.paras
        while i <= pre_end and (remaining > 0 or last) do
          remaining = remaining - block_para_count(blocks[i])
          table.insert(body, blocks[i])
          i = i + 1
        end
        if not piece.drop and #body > 0 then
          local kind = piece.name:match("^(%a+)") or "dedication"
          if kind ~= "dedication" and kind ~= "epigraph" then kind = "dedication" end
          table.insert(out, pandoc.Header(1, pandoc.Str(titlecase(kind)),
            pandoc.Attr("", { "unnumbered", "fm-piece-head" }, { ["epub:type"] = kind })))
          table.insert(out, pandoc.Div(body, pandoc.Attr("", { "fm-piece", "fm-" .. kind })))
        end
      end
      while i <= pre_end do table.insert(out, blocks[i]); i = i + 1 end
    end
  end

  -- H1 sections: drop title/toc sections, tag the rest.
  local h1_index, dropping = 0, false
  while i <= #blocks do
    local b = blocks[i]
    if b.t == "Header" and b.level == 1 then
      h1_index = h1_index + 1
      local sec = book_map.sections[h1_index]
      local kind = sec and sec.kind or "body"
      dropping = (kind == "title" or kind == "toc")
      if not dropping then
        local t = epub_type(pandoc.utils.stringify(b.content), kind)
        b.attr.attributes["epub:type"] = t
        table.insert(out, b)
      end
    elseif not dropping then
      table.insert(out, b)
    end
    i = i + 1
  end
  doc.blocks = out
  return doc
end
