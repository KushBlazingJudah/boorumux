# Filters

Boorumux has support for blacklisting certain tags from ever being seen from the
paginated search view.
It ranges from rudimentary "tag blacklist" to filtering out images based on
their rating and if they contain certain tags.

**Everything here does not apply to search queries; they are passed through
as is onto the booru.**
This is **only** for the client-side blacklist.

The blacklist can be configured in `boorumux.json` in the `blacklist` attribute.

Rating can be selected by using the "rating:..." pseudo-tag.
Valid values:

- `general`, `g`, `sfw`, `safe`
- `questionable`, `q`
- `sensitive`, `s`
- `explicit`, `e`

You can also use a pipe (`|`) to try to match against several ratings.
For example, `rating:q|s|e` would match a questionable post, a sensitive post,
or an explicit post.

Additionally, you can negate any tag by placing a `-` behind it as per usual.

## Examples

Example 1:

```json
[
  "guro",
  "scooby-doo",
  "rating:explicit"
]
```

Example 2, featuring multiple criteria filters:

```json
[
  "lucky_star -chibi",
  "rating:e|s touhou",
]
```

- On posts with the tag `lucky_star`, filter the ones who do not have `chibi`.
- On posts with `rating:explicit` or `rating:sensitive`, filter `touhou`.

Example 3, using a hierarchical style:

```json
{
  "rating:e": [
    "lucky_star",
    "touhou",
    "pokemon"
  ],
  "alternate_costume": {
    "izumi_konata": [
      "swimsuit",
      "halloween"
    ],
    "takara_miyuki": "hat"
  }
}
```

This one may look odd, so here's what the functionally the same list form would
be:

```json
[
  "rating:e lucky_star",
  "rating:e touhou",
  "rating:e pokemon",
  "alternate_costume izumi_konata swimsuit",
  "alternate_costume izumi_konata halloween",
  "alternate_costume takara_miyuki hat",
]
```

Essentially, the key acts as a filter that is applied to everything inside.

As you might've noticed, it is completely acceptable to mix and match the three
valid types.
You can start off with whatever and end with whatever; as long as the JSON
parses fine, and nothing is `null`, a boolean, or a number, it'll work.
