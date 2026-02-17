# Link Manager — Future Features

## V2: Vector Search & AI

### Semantic Search (pgvector)

- Generate embeddings from title + description + generated_notes
- Store in vector column on links table
- Enable similarity search: "find links like this one"
- Hybrid search: combine FTS ranking with vector similarity

### RAG Queries

- "What are my best resources for building a fuzz pedal?"
- Query across all saved links with context
- Return relevant links + synthesized answer

### Improved Generated Notes

- Better extraction from complex pages
- Support for PDFs, YouTube transcripts
- Multiple summary styles (bullet points, paragraph, Q&A)

---

## V2: Enhanced Organization

### Category Card Pagination

- Limit links displayed per category card to ~40 links
- Show total link count on category card header
- Add "View All {count} Links" button at bottom of card to expand/navigate to full view
- Consider implementing lazy loading or modal view for expanded category
- Current: Each CategoryCard fetches all links (limited to 15 in frontend but not enforced)
- Goal: Better performance for categories with hundreds of links while maintaining quick overview

### Nested Categories

```
Electronics/
  ├── Amplifiers/
  │   ├── Tube
  │   └── Solid State
  └── Pedals/
      ├── Overdrive
      └── Fuzz
```

### Smart Collections

- Auto-generated based on tags, stars, recency
- "Highly rated but not visited in 30 days"
- "Recently added, unrated"

### Duplicate Detection

- Warn when adding URL that already exists
- Merge duplicates across projects

### Drag-and-Drop & Custom Ordering

- **Custom Reordering**: Support manual ordering of links within a category using Drag-and-Drop.
- **Cross-Category Moves**: Drag a link from one category card to another to update its category.
- **Lexicographical String Ordering**:
  - Implement ordering using lexicographical strings (similar to Figma or Linear) to allow arbitrary insertions without re-indexing all items.
  - New service/utility specifically for string ordering logic with comprehensive test cases.
  - Insert between `A` and `B` → `AN`, etc.
- **Rebalance Service**:
  - Admin feature to re-normalize/rebalance ordering strings for a category if they become excessively long.
  - Triggerable via UI button or API.

---

## V2: Browser Extension

### Quick Capture

- Click extension icon to save current page
- Popup with project/category selection
- Optional tags and stars
- Auto-fetch title and description

### Context Menu

- Right-click link → "Save to Link Manager"
- Right-click selection → "Search in Link Manager"

### Sync Status

- Show if current page is already saved
- Quick edit existing entry

---

## V3: Collaboration

### Sharing

- Share individual links with expiring URLs
- Share collections/categories publicly
- Embed link lists in other sites

### Teams

- Shared projects within a team
- Permission levels (view, edit, admin)
- Activity feed

### Import/Export

- Import from Pocket, Raindrop, browser bookmarks
- Sync with Obsidian vault
- OPML export for RSS readers

---

## V3: Mobile

### Progressive Web App

- Installable on mobile
- Offline viewing of saved links
- Background sync

### Mobile-First Capture

- Share sheet integration (iOS/Android)
- Quick capture without opening app

---

## V3: Analytics

### Personal Insights

- Reading patterns over time
- Most visited domains
- Tag usage trends
- "Link rot" detection (dead URLs)

### Dashboard

- Visual overview of collection
- Graphs and charts
- Export reports

---

## Technical Debt / Improvements

### Performance

- Pagination everywhere
- Lazy loading for large collections
- CDN for favicons

### Security

- OAuth/SSO integration
- 2FA
- API rate limiting
- Audit logging

### Developer Experience

- OpenAPI spec
- Client SDKs
- Webhook notifications
- Plugin system
