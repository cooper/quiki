# Models

Models are pieces of quiki source code that can be borrowed by multiple pages.
This allows you to create reusable templates for consistency across your wiki
and eliminate repetitious page source code.

* [Models](#models)
  * [Creating models](#creating-models)
  * [Using models](#using-models)

## Creating models

For safety, you cannot include any old page file into another. Instead, models are
stored in a dedicated `models` directory within the wiki root.

Model source files can contain any quiki code, but it is common to use them
in conjunction with [`html{}`](blocks.md#html) or [`format{}`](blocks.md#format)
to make HTML templates. The result is that your actual page files are far less
cluttered, with all the ugly HTML hidden behind a model.

Inside the model source file, the special variable `@m` refers to any key-value
options provided to the model from the main page.

## Using models

quiki has a special syntax for using models. Write them like any block,
except prefix the model name with a dollar sign (`$`).
```
$my_model {
    option1: Something;
    option2: Another option;
}
```
Note: From within the model source, those options can be retrieved with
`@m.option1` and `@m.option2`.

This convenient syntax is the same as writing the long form:
```
model [my_model] {
    option1: Something;
    option2: Another option;
}
```
