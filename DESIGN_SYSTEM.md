# Lawrence UI Design System

This document defines the design system guidelines for the Lawrence OpenTelemetry platform UI.

## Typography

### Font Weight Scale

We use a systematic approach to font weights to establish clear visual hierarchy:

| Weight | Value | Usage | Example Classes |
|--------|-------|-------|-----------------|
| **Light** | 300 | De-emphasized text, secondary information | `font-light` |
| **Regular** | 400 | Body text, default content | `font-normal` |
| **Medium** | 500 | Labels, slight emphasis, nav items | `font-medium` |
| **Semibold** | 600 | Subheadings, card titles, section headers | `font-semibold` |
| **Bold** | 700 | Page headings, primary emphasis | `font-bold` |

### Typography Guidelines

**Headings:**
- H1 (Page Title): `text-3xl font-bold`
- H2 (Section): `text-2xl font-semibold`
- H3 (Card Title): `text-lg font-semibold`
- H4 (Subsection): `text-base font-semibold`

**Body Text:**
- Default: `text-sm font-normal`
- Labels: `text-sm font-medium`
- Helper/Muted: `text-xs text-muted-foreground/70`

**Code/Monospace:**
- IDs: `font-mono text-xs text-muted-foreground`
- Code blocks: `font-mono text-sm`

## Colors

### Status Colors

Status colors communicate state and guide user attention:

| Status | Light Mode | Dark Mode | Meaning |
|--------|-----------|-----------|---------|
| **Online** | Green-600 (`#16a34a`) | Green-400 (`#4ade80`) | Active, healthy, positive |
| **Offline** | Amber-600 (`#d97706`) | Amber-400 (`#fbbf24`) | Attention needed, warning |
| **Error** | Red-500 (`#ef4444`) | Red-500 (`#ef4444`) | Critical, failed, destructive |
| **Unknown** | Gray-500 (`#6b7280`) | Gray-400 (`#9ca3af`) | Neutral, undefined |

### Status Badges

```tsx
// Online
className="bg-green-500/10 text-green-700 border-green-500/20
           dark:bg-green-500/20 dark:text-green-400 dark:border-green-500/30"

// Offline
className="bg-amber-500/10 text-amber-700 border-amber-500/20
           dark:bg-amber-500/20 dark:text-amber-400 dark:border-amber-500/30"
```

## Spacing

### Consistent Units

Use multiples of 4px for consistent spacing:

- **4px** (1): Tight spacing, icon gaps
- **8px** (2): Small gaps between related items
- **12px** (3): Default gap between elements
- **16px** (4): Standard padding for cells, cards
- **24px** (6): Spacing between sections
- **32px** (8): Large section spacing, page margins

### Component Spacing

**Tables:**
- Header height: `h-14` (56px)
- Cell padding: `px-4 py-3` (16px horizontal, 12px vertical)
- Row hover: smooth transition with `duration-200`

**Cards:**
- Padding: `p-4` or `p-6` depending on content density
- Gap between cards: `space-y-4` or `gap-4`

**Buttons:**
- Default height: `h-9` (36px)
- Small: `h-8` (32px)
- Large: `h-10` (40px)

## Depth & Shadows

### Shadow Scale

Create depth and hierarchy using shadows:

```css
/* Subtle elevation */
shadow-xs: 0 1px 2px rgba(0, 0, 0, 0.05)
shadow-sm: 0 1px 3px rgba(0, 0, 0, 0.1)

/* Standard elevation */
shadow: 0 4px 6px rgba(0, 0, 0, 0.1)
shadow-md: 0 6px 12px rgba(0, 0, 0, 0.15)

/* Strong elevation */
shadow-lg: 0 10px 20px rgba(0, 0, 0, 0.2)
shadow-xl: 0 20px 40px rgba(0, 0, 0, 0.25)
shadow-2xl: 0 25px 50px rgba(0, 0, 0, 0.3)
```

### Button States

Buttons should feel tactile with depth changes:

```tsx
// Rest state
shadow-sm

// Hover state
shadow-md (lifting effect)

// Active/pressed state
shadow-sm + scale-[0.98] (pressing effect)
```

## Interactive States

### Button Hierarchy

**Primary Actions:**
- Default variant (filled background)
- Higher visual weight
- Used for main CTAs (Create, Save, Submit)

**Secondary Actions:**
- Ghost or outline variant
- Lower visual weight
- Used for supporting actions (Refresh, Cancel)

**Destructive Actions:**
- Ghost variant with red text at rest
- Red background on hover
- Used for delete, remove operations

### Hover States

**Tables:**
- Row hover: `hover:bg-muted/60` with `transition-all duration-200`

**Buttons:**
- Add shadow increase on hover
- Smooth transitions with `duration-200`

**Cards:**
- Subtle shadow increase: `hover:shadow-md`
- Optional scale: `hover:scale-[1.02]`

## Accessibility

### Focus States

All interactive elements must have visible focus indicators:

```tsx
focus-visible:border-ring
focus-visible:ring-ring/50
focus-visible:ring-[3px]
```

### Color Contrast

- Ensure WCAG 2.1 AA compliance (4.5:1 for normal text, 3:1 for large text)
- Test both light and dark modes
- Use tools like WebAIM Contrast Checker

### Screen Readers

- Include `sr-only` labels for icon-only buttons
- Proper ARIA labels and roles
- Semantic HTML structure

## Best Practices

### Component Consistency

1. **Reuse components** - Use existing components (PageTable, TruncatedId, InfoCard) before creating new ones
2. **Follow patterns** - Maintain consistency with existing component structure
3. **Dark mode support** - Always include dark mode variants for colors
4. **Responsive design** - Ensure layouts work on mobile, tablet, and desktop

### Performance

1. **Smooth transitions** - Use `duration-200` for most transitions
2. **Avoid layout shifts** - Set explicit dimensions where possible
3. **Optimize animations** - Keep animations subtle and performant

### Code Style

1. **Tailwind classes** - Group by type (layout → spacing → colors → states)
2. **Use cn()** - Combine classes with the `cn()` utility
3. **Extract variants** - Use `cva()` for component variants
4. **Document props** - Add JSDoc comments for component props

## Examples

### Creating a New Card Section

```tsx
<Card>
  <CardHeader>
    <CardTitle className="text-lg font-semibold">
      Section Title
    </CardTitle>
  </CardHeader>
  <CardContent className="space-y-3">
    {/* Content here */}
  </CardContent>
</Card>
```

### Displaying a Truncated ID

```tsx
<TruncatedId
  id={agent.id}
  maxLength={8}
  showCopyButton={true}
/>
```

### Status Badge

```tsx
<Badge
  variant="default"
  className="bg-green-500/10 text-green-700 border-green-500/20
             dark:bg-green-500/20 dark:text-green-400 dark:border-green-500/30"
>
  Online
</Badge>
```

## Continuous Improvement

This design system is a living document. When adding new patterns:

1. Document the pattern in this file
2. Add examples to help other developers
3. Ensure consistency with existing patterns
4. Consider accessibility implications
5. Test in both light and dark modes
