# Bastion Document Format

The document format is very simple. It starts with a header containing
`key:value` pairs separated by newlines. Keys can have multiple values by
repeating the key with a distinct value on a separate line.

After the header section is the format specifier. This is three equal signs
followed by a string declaring the format, and followed by three more equal
signs.

Below the format specifier is the content itself. This content will be
interpreted according to the format specifier.

Example:
```
Title: This is the Title of the Document
Tag: Math
=== markdown ===
This is an example of a document. When I provide the math tag, I have access to
mathJax which allows me to write pretty math like so:

$$
x \in \integers

ax^2 + bx + c
$$
```

## Notable Key Value Pairs
Bastion uses some key value pairs from this document format. The following list
describes which keys are used for which purpose.

- Title: Used as the title of the generated article
- Created: Used as the date for when the article was published an for sorting
  it in the index.
- Updated: Contains the time the article was updated. This can appear multiple
  times.
- Pinned: If `true`, then the generated article is pinned to the top of each
  webpage.

## Tags
Bastion is designed to use different tags for different purposes. The table
below describes what they are used for:

Key | Value | Description
-------------------------
Tag | Math  | Loads mathJax to display math equations
