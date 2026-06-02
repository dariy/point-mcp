package prompts

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Register wires the create_landing_page prompt into s.
func Register(s *mcp.Server) {
	s.AddPrompt(&mcp.Prompt{
		Name:        "create_landing_page",
		Description: "Guides the model through the full workflow to author and create a theme-aware immersive landing page as a draft post.",
		Arguments: []*mcp.PromptArgument{
			{
				Name:        "topic",
				Description: "Subject, product, or service the landing page should promote.",
				Required:    true,
			},
			{
				Name:        "media_paths",
				Description: "Optional comma-separated absolute local file paths to upload as hero/section media (e.g. /home/user/hero.jpg,/home/user/feature.png).",
				Required:    false,
			},
		},
	}, createLandingPageHandler)
}

func createLandingPageHandler(_ context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	topic := strings.TrimSpace(req.Params.Arguments["topic"])
	if topic == "" {
		return nil, fmt.Errorf("topic argument is required")
	}
	mediaPaths := strings.TrimSpace(req.Params.Arguments["media_paths"])

	mediaStep := buildMediaStep(mediaPaths)
	body := buildBody(topic, mediaStep)

	return &mcp.GetPromptResult{
		Description: fmt.Sprintf("Create a theme-aware immersive landing page about: %s", topic),
		Messages: []*mcp.PromptMessage{
			{
				Role:    "user",
				Content: &mcp.TextContent{Text: body},
			},
		},
	}, nil
}

func buildMediaStep(mediaPaths string) string {
	if mediaPaths == "" {
		return "If you have local image or video files for the landing page, call **point_upload_media** for each one now and record the returned `path` values. Otherwise skip this step."
	}
	paths := strings.Split(mediaPaths, ",")
	var lines []string
	for _, p := range paths {
		p = strings.TrimSpace(p)
		if p != "" {
			lines = append(lines, "- `"+p+"`")
		}
	}
	if len(lines) == 0 {
		return "If you have local image or video files for the landing page, call **point_upload_media** for each one now and record the returned `path` values. Otherwise skip this step."
	}
	return "Call **point_upload_media** once for each file below. Record the returned `path` (e.g. `/2026/06/hero.jpg`) for use in content and as `thumbnail_path`.\n\n" + strings.Join(lines, "\n")
}

// bt is a backtick, used in raw string templates to avoid the
// "backtick inside raw string literal" restriction.
const bt = "`"

func buildBody(topic, mediaStep string) string {
	// §TOPIC§, §MEDIA§ are substituted below.
	// bt is used everywhere a backtick is needed inside a raw-string segment.
	raw := `You are creating a polished, theme-aware immersive landing page about: **§TOPIC§**

Follow the steps below in order. Do not skip any step.

---

## Step 1 — Discover blog context and theme

Call **point_get_context**, **point_get_theme_css**, and **point_get_syntax_guidelines** in parallel. Record:
- ` + bt + `base_url` + bt + ` (for the preview link)
- ` + bt + `blog_title` + bt + ` and ` + bt + `author_name` + bt + `
- Active theme ` + bt + `preview_color` + bt + ` → maps to CSS variable ` + bt + `--color-accent` + bt + `
- Active theme ` + bt + `has_dark_mode` + bt + ` → if true, use dark-mode–friendly contrast ratios
- Markdown extensions and allowed HTML/CSS from guidelines

---

## Step 2 — Upload media

§MEDIA§

---

## Step 3 — Author the landing page

### Content format rules

Use **markdown** (` + bt + `formatter: "markdown"` + bt + `) as the base. Prefer markdown extensions (attributes, fences) over raw HTML for structural layout.

#### Fenced blocks (structural sections)

Wrap major sections using the ` + bt + `::: {.classname} ... :::` + bt + ` syntax. These render as ` + bt + `<div class="classname">...</div>` + bt + ` after markdown processing.

Example:

` + "```" + `
::: {.hero}
# Welcome

Transform the way your team works.

[Get started](#cta){.btn}
:::

::: {.features}
## Why choose us?

::: {.feature-card}
### Fast
Description here.
:::
:::
` + "```" + `

#### Attributes syntax

Attach classes or IDs to elements like links or images using ` + bt + `{.class #id}` + bt + ` immediately after the element.
Example: ` + bt + `[Link Text](url){.btn}` + bt + ` or ` + bt + `![Alt text](path){.hero-img}` + bt + `

#### Allowed HTML tags (bluemonday policy)

These tags survive sanitization — use them for advanced layouts:

**Text/inline:** ` + bt + `br h1 h2 h3 h4 h5 h6 p span em strong i b u s del ins mark` + bt + `
**Lists:** ` + bt + `ul ol li blockquote code pre hr` + bt + `
**Layout:** ` + bt + `header section div article aside main nav figure figcaption` + bt + `
**Links:** ` + bt + `a` + bt + ` — allowed attrs: ` + bt + `href title target rel` + bt + `
**Images:** ` + bt + `img` + bt + ` — allowed attrs: ` + bt + `src alt title width height loading` + bt + `
**Video:** ` + bt + `video` + bt + ` — allowed attrs: ` + bt + `src controls autoplay muted loop playsinline poster preload width height` + bt + `
**Audio:** ` + bt + `audio source` + bt + ` — allowed attrs: ` + bt + `src type controls autoplay loop preload` + bt + `
**SVG:** ` + bt + `svg g path circle rect line polyline polygon ellipse text tspan` + bt + ` with standard geometry, fill, and stroke attrs

All layout and text elements accept ` + bt + `class` + bt + `, ` + bt + `id` + bt + `, and ARIA attrs (` + bt + `role aria-label aria-hidden aria-labelledby aria-describedby` + bt + `).

#### Allowed inline ` + bt + `style` + bt + ` properties

Safe to use in ` + bt + `style="..."` + bt + ` attributes (all others are stripped):

` + bt + `color background-color background font-size font-weight font-style font-family font-variant text-align text-decoration text-transform text-indent line-height letter-spacing word-spacing margin padding border border-radius border-color border-width border-style width max-width min-width height max-height min-height display flex-direction flex-wrap justify-content align-items align-self flex gap grid-template-columns float clear overflow opacity vertical-align list-style white-space` + bt + `

**Stripped — do not use in style attrs:** ` + bt + `position: fixed` + bt + `, ` + bt + `position: sticky` + bt + `, ` + bt + `z-index` + bt + `, ` + bt + `@import` + bt + `, ` + bt + `url()` + bt + ` with external HTTPS URLs, ` + bt + `content` + bt + `

#### Per-post CSS (` + bt + `css` + bt + ` field)

Provide scoped CSS in the ` + bt + `css` + bt + ` field of ` + bt + `point_create_post` + bt + `. Reference theme variables for coherence with the active theme:

` + "```css" + `
.hero {
  background-color: var(--color-accent, #4f46e5);
  color: #fff;
  padding: 4rem 2rem;
  text-align: center;
}
.hero h1 {
  font-size: 2.5rem;
  font-weight: 700;
  margin-bottom: 1rem;
}
.btn {
  display: inline-block;
  padding: 0.75rem 1.5rem;
  background-color: #fff;
  color: var(--color-accent, #4f46e5);
  border-radius: 0.375rem;
  font-weight: 600;
  text-decoration: none;
}
.features {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 2rem;
  padding: 3rem 2rem;
}
.feature-card {
  padding: 1.5rem;
  border: 1px solid var(--color-accent, #4f46e5);
  border-radius: 0.5rem;
}
` + "```" + `

**CSS restrictions enforced by the server:**
- ` + bt + `position: fixed` + bt + ` and ` + bt + `position: sticky` + bt + ` are stripped (use ` + bt + `position: relative` + bt + ` or ` + bt + `position: absolute` + bt + `)
- ` + bt + `z-index` + bt + ` is stripped
- ` + bt + `@import` + bt + ` is stripped
- ` + bt + `url()` + bt + ` with external HTTPS URLs is stripped (use uploaded media paths instead)
- ` + bt + `content` + bt + ` property is stripped

The ` + bt + `css_warnings` + bt + ` field in the response lists any stripped declarations.

#### Media embedding

Bare media paths on their own line auto-expand to ` + bt + `<img>` + bt + `, ` + bt + `<video>` + bt + `, or ` + bt + `<audio>` + bt + `:
` + "```" + `
/2026/06/hero.jpg
` + "```" + `
Or use explicit markdown: ` + bt + `![alt text](/2026/06/hero.jpg)` + bt + ` or ` + bt + `![alt text](/2026/06/hero.jpg){.hero-img}` + bt + `
` + bt + `<video src="/2026/06/video.mp4" controls></video>` + bt + `

### Suggested landing page structure

Adapt the sections to the topic. A typical structure:

1. **Hero** — headline, sub-headline, CTA button, optional hero image
2. **Features / Benefits** — 2–4 cards or bullet groups with ` + bt + `::: {.features}` + bt + `
3. **Social proof** — blockquote or testimonial in ` + bt + `::: {.testimonial}` + bt + `
4. **Call to action** — prominent button or signup form in ` + bt + `::: {.cta}` + bt + `

---

## Step 4 — Create the post as a draft

Call **point_create_post** with these fields:

| Field | Value |
|---|---|
| ` + bt + `title` + bt + ` | Descriptive page title |
| ` + bt + `content` + bt + ` | Full markdown body from Step 3 |
| ` + bt + `css` + bt + ` | Scoped CSS from Step 3 |
| ` + bt + `formatter` + bt + ` | ` + bt + `"markdown"` + bt + ` |
| ` + bt + `immersive_mode` + bt + ` | ` + bt + `"true"` + bt + ` |
| ` + bt + `status` + bt + ` | ` + bt + `"draft"` + bt + ` |
| ` + bt + `thumbnail_path` + bt + ` | Path of hero image from Step 2 (if any) |
| ` + bt + `excerpt` + bt + ` | One-sentence summary for listings |

If ` + bt + `css_warnings` + bt + ` in the response is non-empty, remove the flagged declarations from your CSS and call **point_update_post** to fix it.

---

## Step 5 — Return the preview link

Call **point_generate_preview_link** with the post ID returned in Step 4. Return the preview URL to the user so they can review the draft before publishing.`

	body := strings.ReplaceAll(raw, "§TOPIC§", topic)
	body = strings.ReplaceAll(body, "§MEDIA§", mediaStep)
	return body
}
