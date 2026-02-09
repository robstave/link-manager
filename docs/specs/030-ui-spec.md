# Link Manager — UI Specification

## Layout Overview

```
┌─────────────────────────────────────────────────────────────────┐
│  Header: Logo | Search | User Menu                              │
├──────────────┬──────────────────────────────────────────────────┤
│              │                                                  │
│   Projects   │              Main Content                        │
│   Sidebar    │                                                  │
│              │   ┌─────────────────┐  ┌─────────────────┐      │
│   • Default  │   │  Category A     │  │  Category B     │      │
│   • Pedals   │   │  ├─ link 1      │  │  ├─ link 1      │      │
│   • Synths   │   │  ├─ link 2      │  │  ├─ link 2      │      │
│   • ...      │   │  └─ (15 max)    │  │  └─ (15 max)    │      │
│              │   │  [View All →]   │  │  [View All →]   │      │
│              │   └─────────────────┘  └─────────────────┘      │
│              │                                                  │
│   ─────────  │   ┌─────────────────┐  ┌─────────────────┐      │
│   Tags       │   │  Category C     │  │  Category D     │      │
│              │   │  ...            │  │  ...            │      │
│   • fuzz     │   └─────────────────┘  └─────────────────┘      │
│   • delay    │                                                  │
│              │                                                  │
└──────────────┴──────────────────────────────────────────────────┘
```

---

## Responsive Behavior

| Viewport | Sidebar | Category Columns |
|----------|---------|------------------|
| Desktop (>1024px) | Fixed left, always visible | 2 columns |
| Tablet (768-1024px) | Collapsible | 1-2 columns |
| Mobile (<768px) | Hamburger menu | 1 column |

---

## Components

### Header

- **Logo**: Click returns to default project
- **Search**: Global search across all user's links
  - Instant results dropdown
  - Enter goes to full search results page
- **User Menu**: Settings, Logout

---

### Projects Sidebar

Tight vertical list of project names.

**Each project row**:
- Project name (truncate if long)
- Link count badge
- Active state highlighting

**Actions**:
- Click → Load project in main content
- Right-click / long-press → Context menu (Rename, Delete, Reorder)
- Bottom: "+ New Project" button

**Tags Section** (below projects):
- Collapsible
- Shows all user's tags
- Click tag → Filter view showing all links with that tag

---

### Main Content — Project View

Shows all categories for selected project as card containers.

**Category Card**:
```
┌─────────────────────────────────────┐
│ 📁 Schematics                  (42) │
├─────────────────────────────────────┤
│ ⭐⭐⭐⭐⭐ 🔗 Amazing Fuzz Schematic  │
│ ⭐⭐⭐⭐  🔗 Tube Screamer Clone     │
│ ⭐⭐⭐   🔗 Big Muff Analysis       │
│ ...                                 │
│ (showing top 15 by stars)           │
├─────────────────────────────────────┤
│ View All →                          │
└─────────────────────────────────────┘
```

**Link Row** (dense):
- Favicon (icon_url or default)
- Stars indicator (visual, e.g., ★★★☆☆ or 8/10)
- Title (clickable, opens in new tab via /click endpoint)
- Hover → Tooltip/popover with more details

**Sorting Options** (per category or global):
- By stars (default)
- By most clicked
- By recently added
- By recently clicked

---

### Link Hover/Popover

On hover over a link row (desktop), show popover with:

```
┌─────────────────────────────────────┐
│ Amazing Fuzz Schematic              │
│ ─────────────────────────────       │
│ ⭐ 8/10  👁 42 clicks               │
│ https://example.com/fuzz            │
│                                     │
│ Great schematic for a silicon       │
│ fuzz with switchable clipping...    │
│ ─────────────────────────────       │
│ Tags: fuzz, silicon, diy            │
│ Last clicked: 2 days ago            │
└─────────────────────────────────────┘
```

Fields shown:
- Title
- Stars
- Click count
- URL
- Description (truncated)
- Tags
- Last clicked

---

### Category View (Full Page)

When clicking "View All" on a category card:

```
┌─────────────────────────────────────────────────────────────────┐
│ ← Back to Pedals                                                │
│                                                                 │
│ Schematics (42 links)                              [+ Add Link] │
│ ─────────────────────────────────────────────────────────────   │
│ Sort: [Stars ▼] [Clicks] [Recent] [Added]                       │
│ ─────────────────────────────────────────────────────────────   │
│                                                                 │
│ ┌─────────────────────────────────────────────────────────────┐ │
│ │ 🔗 Amazing Fuzz Schematic                        ⭐ 8/10    │ │
│ │ https://example.com/fuzz                                    │ │
│ │ Great schematic for a silicon fuzz with switchable...       │ │
│ │ Tags: fuzz, silicon  |  42 clicks  |  Last: 2 days ago      │ │
│ └─────────────────────────────────────────────────────────────┘ │
│                                                                 │
│ ┌─────────────────────────────────────────────────────────────┐ │
│ │ 🔗 Tube Screamer Clone                           ⭐ 7/10    │ │
│ │ ...                                                         │ │
│ └─────────────────────────────────────────────────────────────┘ │
│                                                                 │
│ [Load More...]                                                  │
└─────────────────────────────────────────────────────────────────┘
```

Each link shows:
- Favicon + Title (clickable)
- Stars
- URL
- Description snippet
- Tags (clickable)
- Click count
- Last clicked

**Actions**:
- Click title → Open URL in new tab (tracked)
- Click tag → Filter by tag
- Row menu (⋮) → Edit, Move, Delete, Toggle Cart

---

### Link Detail Modal

Full details when clicking edit or dedicated detail view:

- URL (editable)
- Title (editable)
- Description (editable)
- Stars (1-10 slider or star picker)
- Project + Category (dropdowns to move)
- Tags (tag input with autocomplete)
- Cart toggle
- User Notes (markdown editor)
- Generated Notes (read-only markdown, with "Regenerate" button)
- Click statistics
- Timestamps

---

### Add Link Flow

**Quick Add** (minimal friction):
1. Click "+ Add Link" or keyboard shortcut
2. Paste URL
3. System fetches title + favicon automatically
4. Optionally add description, tags, stars
5. Save (goes to current project/category)

**Bulk Import** (future):
- Paste multiple URLs
- CSV/JSON upload

---

### Cart View

Dedicated view for cart items (accessible from header or sidebar):
- Same layout as category view
- Shows all cart=true links across all projects
- Quick toggle to remove from cart
- Export buttons (JSON, Obsidian)

---

### Search Results

Full-page search with filters:

```
┌─────────────────────────────────────────────────────────────────┐
│ Search: "fuzz circuit"                             [x Clear]    │
│ ─────────────────────────────────────────────────────────────   │
│ Filters: [All Projects ▼] [All Categories ▼] [All Tags ▼]      │
│ ─────────────────────────────────────────────────────────────   │
│ 12 results                                                      │
│                                                                 │
│ (Link cards with highlighted matches)                           │
└─────────────────────────────────────────────────────────────────┘
```

---

## Visual Design

### Color Palette
- Primary: Deep blue or purple
- Secondary: Complementary accent
- Background: Light mode default, dark mode support
- Stars: Gold/yellow

### Typography
- Clean sans-serif (Inter, System UI)
- Monospace for URLs

### Iconography
- Simple line icons
- Favicons fetched from sites
- Fallback icon for missing favicons

### Animations
- Subtle hover transitions
- Smooth panel slides
- Toast notifications for actions
