# Extended Features of Markdown Reader

## Summary

This document contains information about the features enabled and/or customized
beyond the general Commonmark specification provided by the base Goldmark package.
Some of these enabled extension are provided directly by the Goldmark package,
while others are third-party extensions to Goldmark, and some are specific to
this Markdown Reader application.

## Built-In Extensions

The following are extensions provided as part of the Goldmark package and enabled
by default in this application. I have provided links to definitions and examples for
the specific extension's markdown syntax.

### GitHub-Flavored Markdown (GFM)

This application enables the option `extension.GFM` which is a wrapper that enables:

- **`extension.Table`** -- [GitHub-Flavored Markdown: Tables](https://github.github.com/gfm/#tables-extension-)
- **`extension.Strikethrough`** -- [GitHub-Flavored Markdown: Strikethrough](https://github.github.com/gfm/#strikethrough-extension-)
- **`extension.Linkify`** -- [GitHub-Flavored Markdown: Autolinks](https://github.github.com/gfm/#autolinks-extension-)
- **`extension.TaskList`** -- [GitHub-Flavored Markdown](https://github.github.com/gfm/#task-list-items-extension-)

It also enables the following PHP Markdown Extras:

- **`extension.DefinitionList`** -- [PHP Markdown Extra: Definition lists](https://michelf.ca/projects/php-markdown/extra/#def-list)
- **`extension.Footnote`** -- [PHP Markdown Extra: Footnotes](https://michelf.ca/projects/php-markdown/extra/#footnotes)

Finally, it also enables the **`extension.Typographer`** extension to provide typographic
entities like [Smartypants](https://daringfireball.net/projects/smartypants/)
_(e.g., Straight quotes like&nbsp;`"` and `'`) into "curly" quote HTML entities [or opening and closing quotes])_

### Emojis

> [!TODO]
> Add more details about Emoji syntax.

### Attributes Insertion (Limited)

The application will attempt to render custom attributes using the `{attribute}` syntax on some block-level markdown elements.

- **`{.hello}`** ==> Adds `class="hello"` to the affected block's HTML element.
- **`{#hello}`** ==> Adds `id="hello"` to the affected block's HTML element.
- **`{newattr="hello"}`** ==> Adds `newattr="hello"` to the affected block's HTML element.
- These can be grouped together in a single attribute tag separated by spaces ==> `{.hello #myid attr="foo"}`

For headings _(e.g., ATX headings using `#`, `##`, etc.)_ the attribute tag should be added directly at the end of the heading:

```markdown
## My Heading {.hello}
```

Would render as:

```html
<h2 class="hello">My Heading</h2>
```

For most other block-type markdown, the attributes tag should be placed on the next line after the block you want it to
apply to:

```markdown
- List item 1
- List item 2
- List item 3
{#hello}
```

Should render as:

```html
<ul id="hello">
   <li>List item 1</li>
   <li>List item 2</li>
   <li>List item 3</li>
</ul>
```

> [!NOTE]
> This set of attribute extensions can be **unpredictable,** so do don't be surprised if attributes don't always work
> as you expect. They seem to work fine in most of my testing, but can be very picky about placement and
> indentation of other elements **and** extensions.

### Bracketed Spans

The syntax `[some text]{.foo}` can be used for creating inline spans with attributes applied. However, this will
only work on standard (bare) inline text; you cannot wrap inline markdown links or images.

If a link or image is on a line by itself, you can apply the attribute tag, but this will be applied to the `<p>` tag
that the image or picture is contained in (since it's on a line by itself). This is **NOT** a bracketed-span.

### GitHub and Obsidian Alerts/Callouts

The application provides a set of pre-defined Alerts/Callouts using the `> [!NOTE]` syntax
used on GitHub and for Obsidian. The current formatting and coloration of these alerts is
**NOT** guaranteed to exactly match their GitHub/Obsidian counterparts, but they will render
as alerts/callouts when you use them.

> [!TODO]
> Add more details about pre-defined alerts/callouts and alternate syntax.

### Mermaid Charting/Graphing

> [!TODO]
> Add more details about Mermaid.
