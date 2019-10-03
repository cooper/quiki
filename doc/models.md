# Models

The `model{}` block type is a mechanism by which you can include wikifier
source code from another file. This allows you to create reusable templates
for consistency across your wiki, or to eliminate repetitive code in your
page files.

* [Models](#models)
  * [Creating models](#creating-models)
  * [Using models](#using-models)

## Creating models

For safety, you cannot include any page file. Instead, models should be in a
dedicated directory within your wiki root. This directory is determined by the
`@dir.model` configuration value, which defaults to `models`.

Model source files can contain any wikifier code, but it is common to use them
in conjunction with [`html{}`](blocks.md#html) or [`format{}`](blocks.md#format)
to make HTML templates. The result is that your actual page files are far less
cluttered, with all the ugly HTML hidden behind a model.

Inside the model source file, the special variable `@m` refers to any key-value
options provided to the model from the main page.

## Using models

wikifier has a special syntax for using models. Write them like any block,
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
