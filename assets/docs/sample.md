---
Title: Markdown Sample File [YAML]
#// DocType: techdoc
---

<!-- markdownlint-disable MD004 MD018 MD025 MD033 MD034 MD036 MD042 MD046 MD049 MD051 -->

# Markdown Sample File

[goto-end-->](#docbottom)

## Table of Contents

- [Simple Section Test](#simple-section-test-l1)
- [GitHub Flavored Markdown](#github-flavored-markdown-extensions)
- [PHP Markdown Extension](#php-markdown-extensions)
- [Lists](#lists)
- [Blockquotes](#blockquotes)
- [Typographic Extension](#typographic-extension)
- [Links](#links)
- [Inline HTML](#inline-html)
- [Alert/Callout Examples](#alertcallouts-examples)
  + [Primary Callouts](#primary-callouts)
  + [Alias Callouts](#alias-callouts)
  + [Folding Examples](#folding-examples)
  + [Custom Titles](#custom-titles)
  + [CSS Styles](#css-styles)
- [Images](#images)

# Simple Section Test (L1)

This is a level 1 section.

## Simple Section Test (L2) {.test}

This is a level 2 section.

### Simple Section Test (L3) {.test itemid="hello"}

This is a level 3 section
{#foo itemid=hello}

#### Simple Section Test (L4)

This is a level 4 section

##### Simple Section Test (L5)

This is a level 5 section

###### Simple Section Test (L6)

This is a level 6 section

## GitHub Flavored Markdown Extensions

GFM provides several extensions to CommonMark...

### Tables

```markdown
|   ID   |      Value | Description             |
| :----: | ---------: | :---------------------- |
| 001    | Foo        | Foo is foo for all time |
| 002    | Bar        | Bar is nice to hop      |
| 3    | Baz        | Baz is always there     |
```

|   ID   |      Value | Description             |
| :----: | ---------: | :---------------------- |
| 001    | Foo        | Foo is foo for all time |
| 002    | Bar        | Bar is nice to hop      |
| 3    | Baz        | Baz is always there     |

### Strikethrough

Strikethrough uses one or two tilde symbols (`~` or `~~`) to provide strikethrough support for
text. This works the same as `_` and `*` for emphasis delimiters.

```markdown
~~Hi~~ Hello, ~there~ world!

**This is some bold text ~without~ with a strikethrough.**
```

~~Hi~~ Hello, ~there~ world!

**This is some bold text ~without~ with a strikethrough.**

### Autolinks

```text {.nolabel}
- http://example.com
- user@example.com
- www.example.com
- `http://example.com`  `// NOT recognized as an autolink`
- <https://example.com?hello>
- <user+home@example.com>
- example.io  `// NOT recognized as an autolink`
- <example.io>  `// NOT recognized as an autolink`
- <https://example.io>
```

- http://example.com
- user@example.com
- www.example.com
- `http://example.com`  `// NOT recognized as an autolink`
- <https://example.com?hello>
- <user+home@example.com>
- example.io  `// NOT recognized as an autolink`
- <example.io>  `// NOT recognized as an autolink`
- <https://example.io>

### Task List Extension

```text {label="markdown"}
- [ ] Task list item 1
- [ ] Task list item 2
  - [x] Subtask 1 of item 2
  - [ ] Subtask 2 of item 2
- [ ] Task list item 3
```

- [ ] Task list item 1
- [ ] Task list item 2
  - [x] Subtask 1 of item 2
  - [ ] Subtask 2 of item 2
- [ ] Task list item 3

## PHP Markdown Extensions

### Definition Lists

This is a definition list:

```markdown
Apple
:   Pomaceous fruit of plants of the genus Malus in
    the family Rosaceae.

Orange
:   The fruit of an evergreen tree of the genus Citrus.

    This is paragraph two of the definition for Citrus.

This starts a new paragraph outside the definition. Ex esse nulla labore adipisicing exercitation
```

Apple
:   Pomaceous fruit of plants of the genus Malus in
    the family Rosaceae.

Orange
:   The fruit of an evergreen tree of the genus Citrus.

    This is paragraph two of the definition for Citrus.

This starts a new paragraph outside the definition. Ex esse nulla labore adipisicing exercitation
proident est eiusmod. Est aute nulla proident adipisicing laborum eiusmod proident sit tempor ex
aute sunt aliquip dolor.

```markdown
Term 1
Term 2
:   Definition for Terms 1 and 2.

Term 3
:   Definition for Term 3

Another new paragraph outside the definition list.
```

Term 1
Term 2
:   Definition for Terms 1 and 2.

Term 3
:   Definition for Term 3

Another new paragraph outside the definition list.

### Footnotes

This is some text with a footnote.[^100]

This is some other text with another footnote:[^ab]

> [!NOTE]
>
> Footnotes are rendered seqentially with numbers. Each footnote must have a distinct name. That
> name will be used to link footnote references to footnote definitions, but has no effect on the
> numbering of the footnotes. Names can contain any character valid within an **`id`** attribute in HTML.

[^100]: This is the footnote (it will be footnote #1).
[^ab]: This is the other footnote (it will be footnote #2)

## Lists

### Unordered Lists

Basic Unordered List:

```markdown
- List item 1
- List item 2
- List item 3
```

- List item 1
- List item 2
- List item 3
{.bar}

Multi-level Unordered List:

```markdown
* List item 1
  + Subitem 1.1
  + Subitem 1.2
    - Subitem 1.2.1
  + Subitem 1.3
* List item 2
* List item 3
  + Subitem 3.1
```

* List item 1
  + Subitem 1.1
  + Subitem 1.2
    - Subitem 1.2.1
  + Subitem 1.3
* List item 2
* List item 3
  + Subitem 3.1

### Standard (Numbered) Ordered Lists

Simple Numbered List:

```markdown
1. Item 1
2. Item 2
3. Item 3
```

1. Item 1
2. Item 2
3. Item 3

Multi-level Numbered List:

```markdown
1. Item 1
   1. Subitem 1.1
   1. Subitem 1.2
      1. Subitem 1.2.1
   1. [Subitem 1.3]{.fg-green}
1. Item 2
1. Item 3
   1. Subitem 3.1
```

1. Item 1
   1. Subitem 1.1
   1. Subitem 1.2
      1. Subitem 1.2.1
   1. [Subitem 1.3]{.fg-green}
1. Item 2
1. Item 3
   1. Subitem 3.1

### Fancy Lists

Fancy lists (*pattered after Pandoc's implementation*) provide Markdown syntax for using letters
and roman numerals in ordered lists. This feature only works if the `gm-fancy-lists` Goldmark
extension is enabled.

Simple Fancy List with letters:

A. List item A
B. List item B
C. List item C

Multi-level Fancy List:

```markdown
1. List item 1
#. List item 2
   A. Subitem 2.A
   #. Subitem 2.B
      i. Subitem 2.B.i
      ii. Subitem 2.B.ii
      #. Subitem 2.B.iii
   #. Subitem 2.C
#. List item 3
   a. Subitem 3.a
   b. Subitem 3.b
   #. Subitem 3.c
      I. Subitem 3.c.I
      #. Subitem 3.c.II
   #. Subitem 3.d
#. List item 4
#. List item 5
```

1. List item 1
#. List item 2
   A. Subitem 2.A
   #. Subitem 2.B
      i. Subitem 2.B.i
      ii. Subitem 2.B.ii
      #. Subitem 2.B.iii
   #. Subitem 2.C
#. List item 3
   a. Subitem 3.a
   b. Subitem 3.b
   #. Subitem 3.c
      I. Subitem 3.c.I
      #. Subitem 3.c.II
   #. Subitem 3.d
#. List item 4
#. List item 5

## Blockquotes

This is a basic blockquote:

> **IMPORTANT:** This is a basic blockquote. Exercitation velit velit veniam incididunt. Nulla
> labore officia elit laborum commodo incididunt voluptate minim nostrud duis commodo magna dolor
> eu. Veniam consequat laboris ex est aliquip nostrud cillum cupidatat fugiat qui aliquip eiusmod
> elit sint.

This is a nested blockquote:

```markdown
> **NOTE:**
>
> This is some text content for the blockquote.
> > **TIP:**
> >
> > This is a tip inside the NOTE blockquote.
>
> This continues the original NOTE blockquote.
```

> **NOTE:**
>
> This is some text content for the blockquote.
> > **TIP:**
> >
> > This is a tip inside the NOTE blockquote.
>
> This continues the original NOTE blockquote.

## Typographic Extension

```markdown
This is some text with "quoted text" used to ``demonstrate'' what
the 'Typographic' extension does... with quotes.
```

This is some text with "quoted text" used to ``demonstrate'' what the 'Typographic' extension does ... with quotes.

The actual output will often depend on the particular font being used by your browser or application --- but is supposed to change the quotes and other typographic symbols to follow the [smartypants](https://daringfireball.net/projects/smartypants/) style for typographic transformation.

## Links

This internal link points back to the [Table of Contents](#table-of-contents).

This references the [README.md](README.md) for this program.

> This references a [Javascript Script](https://example.com/myjs.js).
>
> (*If HTML Sanitizing is enabled, this will be altered -- the 'href=' attribute will be renamed to 'bad-href=' and a title will be added with an explanation*).

[//]: # (This behaves like a comment)

> This references a [Python Script](https://example.com/?p=mypy.py&arg1=2).
>
> (*If HTML Sanitizing is enabled, this will be altered -- the 'href=' attribute will be renamed to 'bad-href=' and a title will be added with an explanation*).

## Inline HTML

When enabled in Goldmark, you can include HTML inside the Markdown (*this is disabled in Goldmark by default and has to be enabled explicitly*).

```markdown
<p><a href="https://www.djseli.net/">My Internal Homepage</a></p>
```

<p><a href="https://www.djseli.net/">My Internal Homepage</a></p>

The following is a simple HTML Form:

<form>
  <!-- this is a form section -->
  <label for="fname">First name:</label><br>
  <input type="text" id="fname" name="fname" value="John"><br>
  <label for="lname">Last name:</label><br>
  <input type="text" id="lname" name="lname" value="Doe"><br><br>
  <input type="submit" value="Submit">
</form>

(_If HTML Sanitizing is enabled, you won't see **anything** above this -- the entire `<form></form>` element and all its contents will be removed_)

(_Elements disallowed by HTML Sanitizer: `<script>`, `<dialog>`, `<embed>`, `<iframe>`, `<form>`, `<button>`, `<input>`, `<select>`_)

## Alert/Callouts Examples

This section demonstrates all available callouts in the GFM Plus icon set, including both primary callouts and their aliases.

It also shows examples of callouts with custom titles and folding.

-----

### Primary Callouts

These are the main callout types with dedicated icons in the GFM Plus set.

#### Note

```markdown
> [!NOTE]
> This is a note callout with informational content.
```

> [!NOTE]
> This is a note callout with informational content.

#### Tip

```markdown
> [!TIP]
> This is a tip callout with helpful suggestions.
```

> [!TIP]
> This is a tip callout with helpful suggestions.

#### Important

```markdown
> [!IMPORTANT]
> This is an important callout highlighting crucial information.
```

> [!IMPORTANT]
> This is an important callout highlighting crucial information.

#### Warning

```markdown
> [!WARNING]
> This is a warning callout about potential issues.
```

> [!WARNING]
> This is a warning callout about potential issues.

#### Caution

```markdown
> [!CAUTION]
> This is a caution callout for dangerous situations.
```

> [!CAUTION]
> This is a caution callout for dangerous situations.

#### Bug

```markdown
> [!BUG]
> This is a bug callout for reporting issues or problems.
```

> [!BUG]
> This is a bug callout for reporting issues or problems.

#### Example

```markdown
> [!EXAMPLE]
> This is an example callout demonstrating usage patterns.
```

> [!EXAMPLE]
> This is an example callout demonstrating usage patterns.

#### Failure

```markdown
> [!FAILURE]
> This is a failure callout indicating something went wrong.
```

> [!FAILURE]
> This is a failure callout indicating something went wrong.

#### Question

```markdown
> [!QUESTION]
> This is a question callout for inquiries or help requests.
```

> [!QUESTION]
> This is a question callout for inquiries or help requests.

#### Quote

```markdown
> [!QUOTE]
> This is a quote callout for citations and references.
```

> [!QUOTE]
> This is a quote callout for citations and references.

#### Scroll

```markdown
> [!SCROLL]
> This is a scroll callout for historical content or long-form text.
```

> [!SCROLL]
> This is a scroll callout for historical content or long-form text.

#### Success

```markdown
> [!SUCCESS]
> This is a success callout indicating completion or achievement.
```

> [!SUCCESS]
> This is a success callout indicating completion or achievement.

#### Summary

```markdown
> [!SUMMARY]
> This is a summary callout providing overview information.
```

> [!SUMMARY]
> This is a summary callout providing overview information.

#### Todo

```markdown
> [!TODO]
> This is a todo callout for task lists and action items.
```

> [!TODO]
> This is a todo callout for task lists and action items.

-----
[top](#)

### Alias Callouts

These are alternative names that reference the primary callouts above.

#### Notes (alias for Note)

```markdown
> [!NOTES]
> This uses the "notes" alias but renders as a note callout.
```

> [!NOTES]
> This uses the "notes" alias but renders as a note callout.

#### Info (alias for Note)

```markdown
> [!INFO]
> This uses the "info" alias but renders as a note callout.
```

> [!INFO]
> This uses the "info" alias but renders as a note callout.

#### Information (alias for Note)

```markdown
> [!INFORMATION]
> This uses the "information" alias but renders as a note callout.
```

> [!INFORMATION]
> This uses the "information" alias but renders as a note callout.

#### Tips (alias for Tip)

```markdown
> [!TIPS]
> This uses the "tips" alias but renders as a tip callout.
```

> [!TIPS]
> This uses the "tips" alias but renders as a tip callout.

#### Hint (alias for Tip)

```markdown
> [!HINT]
> This uses the "hint" alias but renders as a tip callout.
```

> [!HINT]
> This uses the "hint" alias but renders as a tip callout.

#### Hints (alias for Tip)

```markdown
> [!HINTS]
> This uses the "hints" alias but renders as a tip callout.
```

> [!HINTS]
> This uses the "hints" alias but renders as a tip callout.

#### Warn (alias for Warning)

```markdown
> [!WARN]
> This uses the "warn" alias but renders as a warning callout.
```

> [!WARN]
> This uses the "warn" alias but renders as a warning callout.

#### Warnings (alias for Warning)

```markdown
> [!WARNINGS]
> This uses the "warnings" alias but renders as a warning callout.
```

> [!WARNINGS]
> This uses the "warnings" alias but renders as a warning callout.

#### Attention (alias for Warning)

```markdown
> [!ATTENTION]
> This uses the "attention" alias but renders as a warning callout.
```

> [!ATTENTION]
> This uses the "attention" alias but renders as a warning callout.

#### Danger (alias for Caution)

```markdown
> [!DANGER]
> This uses the "danger" alias but renders as a caution callout.
```

> [!DANGER]
> This uses the "danger" alias but renders as a caution callout.

#### Error (alias for Caution)

```markdown
> [!ERROR]
> This uses the "error" alias but renders as a caution callout.
```

> [!ERROR]
> This uses the "error" alias but renders as a caution callout.

#### Errors (alias for Caution)

```markdown
> [!ERRORS]
> This uses the "errors" alias but renders as a caution callout.
```

> [!ERRORS]
> This uses the "errors" alias but renders as a caution callout.

#### Fail (alias for Failure)

```markdown
> [!FAIL]
> This uses the "fail" alias but renders as a failure callout.
```

> [!FAIL]
> This uses the "fail" alias but renders as a failure callout.

#### Missing (alias for Failure)

```markdown
> [!MISSING]
> This uses the "missing" alias but renders as a failure callout.
```

> [!MISSING]
> This uses the "missing" alias but renders as a failure callout.

#### Questions (alias for Question)

```markdown
> [!QUESTIONS]
> This uses the "questions" alias but renders as a question callout.
```

> [!QUESTIONS]
> This uses the "questions" alias but renders as a question callout.

#### FAQ (alias for Question)

```markdown
> [!FAQ]
> This uses the "faq" alias but renders as a question callout.
```

> [!FAQ]
> This uses the "faq" alias but renders as a question callout.

#### FAQs (alias for Question)

```markdown
> [!FAQS]
> This uses the "faqs" alias but renders as a question callout.
```

> [!FAQS]
> This uses the "faqs" alias but renders as a question callout.

#### Help (alias for Question)

```markdown
> [!HELP]
> This uses the "help" alias but renders as a question callout.
```

> [!HELP]
> This uses the "help" alias but renders as a question callout.

#### Quotes (alias for Quote)

```markdown
> [!QUOTES]
> This uses the "quotes" alias but renders as a quote callout.
```

> [!QUOTES]
> This uses the "quotes" alias but renders as a quote callout.

#### Cite (alias for Quote)

```markdown
> [!CITE]
> This uses the "cite" alias but renders as a quote callout.
```

> [!CITE]
> This uses the "cite" alias but renders as a quote callout.

#### Citation (alias for Quote)

```markdown
> [!CITATION]
> This uses the "citation" alias but renders as a quote callout.
```

> [!CITATION]
> This uses the "citation" alias but renders as a quote callout.

#### Citations (alias for Quote)

```markdown
> [!CITATIONS]
> This uses the "citations" alias but renders as a quote callout.
```

> [!CITATIONS]
> This uses the "citations" alias but renders as a quote callout.

#### History (alias for Scroll)

```markdown
> [!HISTORY]
> This uses the "history" alias but renders as a scroll callout.
```

> [!HISTORY]
> This uses the "history" alias but renders as a scroll callout.

#### TL;DR (alias for Scroll)

```markdown
> [!TLDR]
> This uses the "tldr" alias but renders as a scroll callout.
```

> [!TLDR]
> This uses the "tldr" alias but renders as a scroll callout.

**_This may not be the way you want this title to be used, so you can use Custom Titles instead..._**

```markdown
> [!TLDR] tl;dr
> This uses the "tldr" alias as before but uses the custom title of 'tl;dr' instead.
>
> (*see [Custom Titles](#custom-titles) for more examples of custom titles*)
```

> [!TLDR] tl;dr
> This uses the "tldr" alias as before but uses the custom title of 'tl;dr' instead.
>
> (*see [Custom Titles](#custom-titles) for more examples of custom titles*)

#### Check (alias for Success)

```markdown
> [!CHECK]
> This uses the "check" alias but renders as a success callout.
```

> [!CHECK]
> This uses the "check" alias but renders as a success callout.

#### Done (alias for Success)

```markdown
> [!DONE]
> This uses the "done" alias but renders as a success callout.
```

> [!DONE]
> This uses the "done" alias but renders as a success callout.

#### Abstract (alias for Summary)

```markdown
> [!ABSTRACT]
> This uses the "abstract" alias but renders as a summary callout.
```

> [!ABSTRACT]
> This uses the "abstract" alias but renders as a summary callout.

#### Abstracts (alias for Summary)

```markdown
> [!ABSTRACTS]
> This uses the "abstracts" alias but renders as a summary callout.
```

> [!ABSTRACTS]
> This uses the "abstracts" alias but renders as a summary callout.

#### Overview (alias for Summary)

```markdown
> [!OVERVIEW]
> This uses the "overview" alias but renders as a summary callout.
```

> [!OVERVIEW]
> This uses the "overview" alias but renders as a summary callout.

#### Overviews (alias for Summary)

```markdown
> [!OVERVIEWS]
> This uses the "overviews" alias but renders as a summary callout.
```

> [!OVERVIEWS]
> This uses the "overviews" alias but renders as a summary callout.

#### Todos (alias for Todo)

```markdown
> [!TODOS]
> This uses the "todos" alias but renders as a todo callout.
```

> [!TODOS]
> This uses the "todos" alias but renders as a todo callout.

#### Todolist (alias for Todo)

```markdown
> [!TODOLIST]
> This uses the "todolist" alias but renders as a todo callout.
```

> [!TODOLIST]
> This uses the "todolist" alias but renders as a todo callout.

#### Task (alias for Todo)

```markdown
> [!TASK]
> This uses the "task" alias but renders as a todo callout.
```

> [!TASK]
> This uses the "task" alias but renders as a todo callout.

#### Tasks (alias for Todo)

```markdown
> [!TASKS]
> This uses the "tasks" alias but renders as a todo callout.
```

> [!TASKS]
> This uses the "tasks" alias but renders as a todo callout.

#### Tasklist (alias for Todo)

```markdown
> [!TASKLIST]
> This uses the "tasklist" alias but renders as a todo callout.
```

> [!TASKLIST]
> This uses the "tasklist" alias but renders as a todo callout.

#### Checklist (alias for Todo)

```markdown
> [!CHECKLIST]
> This uses the "checklist" alias but renders as a todo callout.
```

> [!CHECKLIST]
> This uses the "checklist" alias but renders as a todo callout.

#### Punchlist (alias for Todo)

```markdown
> [!PUNCHLIST]
> This uses the "punchlist" alias but renders as a todo callout.
```

> [!PUNCHLIST]
> This uses the "punchlist" alias but renders as a todo callout.

#### Outline (alias for Todo)

```markdown
> [!OUTLINE]
> This uses the "outline" alias but renders as a todo callout.
```

> [!OUTLINE]
> This uses the "outline" alias but renders as a todo callout.

#### Outlines (alias for Todo)

```markdown
> [!OUTLINES]
> This uses the "outlines" alias but renders as a todo callout.
```

> [!OUTLINES]
> This uses the "outlines" alias but renders as a todo callout.

-----
[top](#)

### Folding Examples

You can also use folding functionality with any of these callouts (*if Folding is enabled, which it
is by default*). The extension looks for a `+` or `-` directly after the `]` and if it finds one of
them it will use the `<details>` and `<summary>` elements to render foldable content. If you do not
include one of these two symbols, the extension will treat the callout as a standard non-foldable
callout and revert to using `<div>` elements (*even if Folding is enabled*).

If Folding is **disabled** then all callouts will be treated the same (*i.e., the `+` or `-` are ignored*).

#### Closed by Default

```markdown
> [!TIP]-
> This tip callout is closed by default due to the minus sign.
```

> [!TIP]-
> This tip callout is closed by default due to the minus sign.

#### Open by Default (Explicit)

```markdown
> [!IMPORTANT]+
> This important callout is explicitly marked as open by default with the plus sign.
```

> [!IMPORTANT]+
> This important callout is explicitly marked as open by default with the plus sign.

-----
[top](#)

### Custom Titles

All callouts support custom titles.

#### Use Existing Icon with Custom Title

If you use a recognized icon entry, it will use the icon configured for that entry and the custom text for the title.

```markdown
> [!SUCCESS] Mission Accomplished
> You can override the default title with any custom text.
```

> [!SUCCESS] Mission Accomplished
> You can override the default title with any custom text.

#### The Custom Title is Rendered As-Is

The title is rendered as-is -- case selection in the custom title is unchanged by the extension!

```markdown
> [!SUCCESS] MiSsIoN AcCoMpLiShEd
> You can override the default title with any custom text.
```

> [!SUCCESS] MiSsIoN AcCoMpLiShEd
> You can override the default title with any custom text.

#### Unknown Callout

If you use an unknown callout name (in this example 'FOO'), the extension will use 'Foo' as the
callout title and set the `data-callout="foo"` attribute. If your CSS does not have an entry for
`[data-attribute="foo"]` it will use whatever default style you configured in your CSS (*see [CSS
Styles](#css-styles) below*).

```markdown
> [!FOO]
> You can use an unrecognized entry for the callout.
```

> [!FOO]
> You can use an unrecognized entry for the callout.

#### No-Icon Callout Style

If you want to create a callout with no icon, you can use the callout name 'noicon' (*or 'none', 'nil' or
'null'*). If you don't supply a custom title, it will render the title using whichever of these
recognized 'noicon' variants you used.

```markdown
> [!NoIcon]
> This is an icon-less callout.
```

> [!NoIcon]
> This is an icon-less callout.

_**This is probably NOT what you want, so you should use a Custom Title instead:**_

```markdown
> [!NOICON] FooBar
> This is an icon-less callout with the title 'FooBar'. Since 'FooBar' is not a recognized callout
> name, it will use the default style (*in this case using blue*).
```

> [!noicon] FooBar
> This is an icon-less callout with the title 'FooBar'. Since 'FooBar' is not a recognized callout
> name, it will use the default style (*in this case using blue*).

#### No Icon Callout with Recognized Custom Title

```markdown
> [!NoIcon] Warning
> This creates a Warning Callout without the Warning Icon, but will be styled using `data-callout="warning"`
> rather than the default styling, because 'warning' is a defined callout name.
```

> [!NoIcon] Warning
> This creates a Warning Callout without the Warning Icon, but will be styled using
> `data-callout="warning"` rather than the default styling, because 'warning' is a defined callout name.

-----
[top](#)

### CSS Styles

_Example CSS snippet for Alert/Callouts..._

```css
/* This applies to all callouts first as the default */
.callout[data-callout] {
   color: blue;
   background-color: transparent;
   /* ...rest of this entry... */
}
/* Specific data-callout values allow you to set styles for each callout kind) */
.callout[data-callout^="note"] {
   color: blue;
   /* ...rest of this entry... */
}
.callout[data-callout^="tip"] {
   color: cyan;
}
```

## Images

![MSU Logo][image1]
MSU Logo (_using a PNG embed in the Markdown document_)

<a id="docbottom"></a>

[image1]: data:image/png;base64,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
