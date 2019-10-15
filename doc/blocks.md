# Blocks

This is the list of built-in block types. For a syntactical explanation of
blocks, see [Language](language.md#blocks).

* [Blocks](#blocks)
  * [clear{}](#clear) - CSS clear
  * [code{}](#code) - block of code
  * [fmt{}](#fmt) - embedded HTML with formatted text
  * [map{}](#map) - key-value datatype
  * [history{}](#history) - table of chronological events
  * [html{}](#html) - embedded HTML
  * [image{}](#image) - image
  * [imagebox{}](#imagebox) - image with border, link, and description
  * [infobox{}](#infobox) - table summarizing article information
  * [invisible{}](#invisible) - silences all blocks inside it
  * [list{}](#list) - list datatype and visible unordered list
  * [model{}](#model) - quiki templates
  * [olist{}](#olist) - ordered list
  * [p{}](#p) - paragraph
  * [sec{}](#sec) - article section with optional header
  * [style{}](#style) - use CSS with quiki

## clear{}

Creates an empty `<div>` with `clear: both`.

```
clear{}
```

This is mostly useless now that the built-in quiki classes `clear`, `clear-left`,
and `clear-right` can be used on any block:

```
p.clear {
    This will clear both sides.
}

p.clear-left {
    This will clear the left side.
}

p.clear-right {
    This will clear the right side.
}
```

## code{}

Wraps a block of code.

The contents is not formatted. When used with [brace escape](language.md#escapes)
(extra `{` and `}`, e.g. `code{{ }}`), there is no need to escape curly brackets
within the code, as long as there is a closing bracket corresponding to each
opening bracket. So, you can paste code right in.

quiki detects the indentation depth of the code based on the first line and
preserves it.

```
code {{
    someCode();
    function someCode() {
        return true;
    }
}}
```

A language may be specified as the name of the block for syntax
highlighting with [code-prettify](https://github.com/google/code-prettify).
Look there for a list of supported languages.

```
code [perl] {{
    $_ = "wftedskaebjgdpjgidbsmnjgc";
    tr/a-z/oh, turtleneck Phrase Jar!/; print;    
}}
```

## fmt{}

Like [`html{}`](#html), except that text formatting is permitted. Often
used with [models](#model).

```
fmt {
    <div>
        This is some HTML.
        [b]quiki formatting[/b] is allowed, such as [@variables].
    </div>
}
```

## map{}

An *ordered* key-value map data type.

Many other block types are built atop map to accept key-value pairs.

Yields no HTML.

```
map {
    key:        value;
    other key:  another value;
}
```

**Shorthand**. A nameless block is assumed to be a map:

```
@myMap: {
    key:        value;
    other key:  another value;
};
```

**Syntax for pairs**. Each pair is separated by a colon (`:`). The left side is
the key; the right is the value. The key must be plain text (no formatting
permitted). Colons (`:`) can be included in the key by prefixing them with the
escape character (`\`). The value may contain
[formatted text](language.md#text-formatting) or a single
[block](language.md#blocks) but not both together. Values are terminated
with the semicolon (`;`). If the value is text, semicolons can be included by
prefixing them with the escape character (`\`). Colons (`:`) do not need to be
escaped in the value.

```
map {
    /* the value of this pair is a block */
    key: list {
        a;
        b;
        c;
    };

    /* the value of this pair is text with an escaped semicolon */
    other key:  another value\; this one has a semicolon;
}
```

**Syntax for standalone values**. In addition to pairs, maps can contain
anonymous values. These are values which are not mapped to a key. Like values in
pairs, standalone values are terminated by the semicolon (`;`). If the value is
text, it should be prefixed with a colon (`:`). The advantage of this is that
colons consequently need not be escaped in the value. For blocks, this prefixing
colon is not required. Semicolons (`;`) can be included in standalone text
values by prefixing them with the escape character (`\`).

```
infobox [My Article] {

    /* this anonymous value is a block */
    image {
        file: mypic.jpg;
    };

    /* this anonymous value is text, so it should be prefixed with a colon */
    :This is some text.;
}
```

**Attributes**. Maps and all map-based block types support attribute fetching
and assignment, which allows you to retrieve and set their values using the
quiki variable attribute syntax.
```
/* define the infobox in a variable so we can access attributes */
@person: infobox [Britney Spears] {
    First name:     Britney;
    Last name:      Spears;
    Age:            35;
};

/* display the infobox */
{@person}

/* access attributes from it elsewhere
   btw this works for all map-based block types */
Did you know that [@person.First_name] [@person.Last_name] is
[@person.Age] years old?
```

**Duplicate keys**. As some block types such as [`infobox{}`](#infobox) and
[`history{}`](#history) use the pairs of a map to display table rows in the
generated HTML, keys may be duplicated. When using a key more than once, both
pairs will be displayed in the resulting HTML, but because only one value can be
associated with each key internally, duplicate keys are suffixed with `_n` where
`n` is incremented for each occurrence, starting at 2.
```
infobox {
    Name:   Britney;    /* shows Name, internally Name   */
    Name:   Spears;     /* shows Name, internally Name_2 */
}
```

**Key fixing**. The original characters of a key will be displayed, but the
internal key will be fixed by replacing all non-word characters with an
underscore (`_`).
```
infobox {
    First name:     Britney;    /* shows First name, internally First_name    */
    First name:     Brittany;   /* shows First name, internally First_name_2  */
    Last name:      Spears;     /* shows Last name, internally Last_name      */
    Last/name:      Speers;     /* shows Last name, internally Last_name_2    */
}
```

## history{}

Displays a timeline of chronological events in a table.

```
history {
    1900: A new century began.;
    2000: A new millennium began.;
}
```

## html{}

Used to embed some HTML.

The contents are not formatted. If formatting is desired,
use [`fmt{}`](#fmt).

```
html {{
    <div>
        This is HTML.
    </div>
}}
```

## image{}

Image.

Typically for embedding standalone images with a nice
border and optional caption you should use [`imagebox{}`](#imagebox) instead.
However, `image{}` is often used inside other block types.

```
infobox [Planet Earth] {
    image {
        file: planet-earth.jpg;
        desc: Earth from space;
    };
    Type:       Planet;
    Population: 23 billion;
    Galaxy:     Milky Way;
}
```

**Options**
* __file__ - _required_, filename of the full-size image.
* __width__ - image width in pixels.
* __height__ - image height in pixels.
* __align__ - `left` or `right` to specify which side of the container the
  image should clear. defaults to `right`.
* __link__ - hyperlink for the image. all [link types](language.md#links) are
  supported, including pages, categories, external wiki links, and external
  site links. `none` is also accepted. defaults to the full-sized image.
* __float__ - alias for __align__.

If neither __width__ nor __height__ is specified, the image will be full-size,
unless its size is constrained by a container. In the above
[`infobox{}`](#infobox) example, the image size is automatically constrained by
the width of the infobox, so dimensions do not need to be specified.

## imagebox{}

Embeds an image with a border and optional caption.

It will be either left or right aligned. Links to the full-sized image.

```
imagebox {
    file:   planet-earth.jpg;
    width:  300px;
    float:  right;  /* this is default */
    desc:   Earth from space;
}
```

**Options**
* __file__ - _required_, filename of the full-size image.
* __width__ - image width in pixels.
* __height__ - image height in pixels.
* __float__ - `left` or `right` to specify which side of the container the
  imagebox should clear. defaults to `right`.
* __align__ - alias for float.

If neither __width__ nor __height__ is specified, the image will be full-size,
unless its size is constrained by a container.

## infobox{}

Displays a summary of information for an article.

Usually there is just one per article, and it occurs before the first section.

```
@page.title: Earth;

infobox [Planet Earth] {

    /* many infoboxes start with an image of the subject.
       the image dimensions are dictated by the infobox  */
    image {
        file: planet-earth.jpg;
    };

    /* standalone text should be prefixed with colon */
    :Earth from space;

    Type:           [[ Planet ]];
    Population:     23 billion;
    Galaxy:         Milky Way;
}

[b]Earth[/b] is the planet on which we live.
```

You can organize the information in sections with `infosec{}`. Each can
optionally have a title as well. The block type can be omitted within
`infobox{}`.

```
@page.title: United States;

infobox [United States of America] {

    Capital:        [! Washington, DC !];  
    Largest city:   [! New York City !];    

    [Goverment] {
        :[! Federal presidential constitutional republic | Republic !];   
        President:              [! Donald Trump !];
        Vice President:         [! Mike Pence !];
        Speaker of the House:   [! Paul Ryan !];
        Chief Justice:          [! John Roberts !];
    };

    [Independence[nl]from [! Great Britain !]] {
        Declaration:            July 4, 1776;
        Confederation:          March 1, 1781;
        Treaty of Paris:        September 3, 1783;
        Constitution:           June 21, 1788;
        Last polity admitted:   March 24, 1976;
    };
}

[b]The United States[/b] ([b]USA[/b], [b]US[/b], [b]America[/b], officially
[b]the United States of America[/b]) is the country in which 100% of
American residents live.
```

## invisible{}

Silences whatever's inside.

```
invisible {
    p {
        Normally this would show a paragraph.
        But since it's inside an invisible block, it shows nothing.
    }
}
```

## list{}

A list datatype.

It may be used by other block types. By itself though,
it yields an unordered list element (`<ul>`).

```
list {
    Item one;
    Item two;
    Item three has an escaped semicolon\; I believe;
}
```

**Syntax**. A value may contain
[formatted text](language.md#text-formatting) or a single
[block](language.md#blocks) but not both simultaneously. Values are terminated
by the semicolon (`;`). If the value is text, additional semicolons may be
included by prefixing them with the escape character (`\`).
```
list {
    Item one;
    Item two;
    Item three can have an escaped semicolon\; I think;
}
```

**Attributes**. Lists support attribute fetching and assignment, which allows
you to retrieve and set their values using the quiki variable attribute
syntax.
```
@alphabet: list {
    a;
    b;
    c;
    ... the rest;
};

Breaking News: [@alphabet.0] is the first letter of the alphabet,
and [@alphabet.25] is the last.
```

See also [`olist{}`](#olist).

## model{}

Allows you to embed a template. See [Models](models.md).

## olist{}

The same as [`list{}`](#list) in all ways, except items are numbered to
indicate a specific order.

## p{}

Paragraph.

You can call it `paragraph{}` or `p{}`. Or you can
call it nothing because stray text within `sec{}` blocks and at document level is
broken down by blank lines into paragraphs.

```
sec [My Section] {
    p {
        A paragraph.
    }
    p {
        Another.
    }
}
```

Same as
```
[My Section] {
    A paragraph.

    Another.
}
```

## sec{}

Section.

You can organize the content of an article by dividing it into sections.
Sections typically have headings, except for the first one, which is
considered the article introduction and uses the page name for the heading.

```
sec {
    This is my intro section. No need to put a title here, since the page title
    will be displayed atop this section.
}

sec [Info] {
    Here we go. This one has a title.

    By the way, a blank line starts a new paragraph.
}

sec [More stuff] {
    You can also put sections inside each other.

    sec [Little header] {
        This section will have a smaller header, since it is nested deeper
        than the top-level sections.
    }
}
```

quiki assumes document-level stray text to be in a section. The block type
`sec` may also be omitted for sections that have headings. So, the above example 
can be more conveiently written as:
```
This is my intro section. No need to put a title here, since the page title
will be displayed atop this section.

[Info] {
    Here we go. This one has a title.

    By the way, a blank line starts a new paragraph.
}

[More stuff] {
    You can also put sections inside each other.

    [Little header] {
        This section will have a smaller header, since it is nested deeper
        than the top-level sections.
    }
}
```

## style{}

Allows you to use CSS with quiki.

See [Styling](styling.md).

```
imagebox {
    file:   some-image.png;
    width:  500px;
    style {
        padding: 5px;
        background-color: red;
    }
}
```
