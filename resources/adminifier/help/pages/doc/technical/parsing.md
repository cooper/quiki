# Parsing

The quiki source language is parsed hierarchically.

## Parsing stages

The parsing process is divided into stages in the following order.

1. [__Master parser__](#master-parser): Data is parsed character-by-character to
separate it into several blocks. Variable definitions are handled. Comments are
stripped. Anything within a block (besides comments and other blocks) is
untouched by the master parser.

2. [__Block parsers__](../blocks.md): Each block type implements its own parser
which parses the data within the block. Block types can be hereditary, in which
case they may rely on another block type for parsing. [Map](../blocks.md#map) and
[List](../blocks.md#list) are the most common parent block types.

3. [__Formatting parser__](../language.md#text-formatting): Many block parsers make
use of a formatting parser afterwards, the one which converts text formatting
such as `[b]` and `[i]` to bold and italic text, etc. Values in
[variable assignment](../language.md#assignment) are also formatted.

## Master parser

The master parser is concerned only with the most basic syntax:
* Dividing the source into [blocks](../language.md#blocks)
* Stripping [comments](../language.md#comments)
* [Variable assignment](../language.md#assignment)
* [Conditionals](../language.md#conditionals)