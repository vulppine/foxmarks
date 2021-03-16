# Implementation Progress

See the [CommonMark specification](https://spec.commonmark.org/0.29/) for more information.

Currently implemented (may or may not abide to specification):

 * Block objects:
   * Paragraphs
   * Headers (ATX-style)
   * Headers (setext-style)
   * Unordered lists
     - Does not abide to specification as this isn't a container block,
       it's just a regular block with the ability to nest
     - Big TODO: Figure out how to make a generic mechanism for a container
       block so that nesting blocks in containers isn't a hassle
   * Horizontal rules
   * Code blocks
     * Indented
     * Fenced

 * Inline objects:
   * Strong text
   * Emphasized text
   * Code spans
   * Links
   * Images

Currently **not** implemented:

 * Block objects:
   * Ordered lists
   * Link references
   * Blockquotes
   * HTML tags (according to specification)
