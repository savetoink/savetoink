# Tag Management Implementation Plan

## Overview

Currently, users can only add tags when creating a new article. The backend already supports full tag operations, but the frontend lacks:
1. Tag management UI (add/edit/remove tags after creation)
2. Tag filtering in the article list
3. API client methods for tag operations
4. Interactive tag components

## Backend API Support (Already Implemented)

The backend provides the following tag-related endpoints:

- `POST /v1/articles/{id}/tags` - Add tags to article
- `PUT /v1/articles/{id}/tags` - Set all article tags (replace)
- `GET /v1/articles/{id}/tags` - Get article tags
- `DELETE /v1/articles/{id}/tags` - Remove specific tags from article
- `GET /v1/tags` - Get all unique tags for the authenticated user
- `GET /v1/articles?tag={tag}` - List articles filtered by tag

### Tag Rules (Enforced by Backend)

- Maximum 10 tags per article
- Maximum 50 characters per tag
- Tags are normalized (lowercase, trimmed, deduplicated)
- Tags can be set to empty array to remove all tags

---

## Phase 1: Shared Library (`@savetoink/shared`)

### Step 1.1: Export Tag Types

**File:** `frontend/shared/src/types/schemas.ts`

Add exports for tag-related types:

```typescript
export type TagsRequest = components["schemas"]["TagsRequest"];
export type TagsResponse = components["schemas"]["TagsResponse"];
```

**Why:** These types are defined in the generated OpenAPI types but not exported for use in the frontend.

**Risks:** None - just adding type exports.

---

### Step 1.2: Update ApiClient Interface and Implementation

**File:** `frontend/shared/src/lib/apiClient.ts`

#### Update `ApiClient` interface:

```typescript
export interface ApiClient {
  // ... existing methods

  getArticles(
    params: {
      page?: number;
      page_size?: number;
      favorite?: boolean;
      tag?: string;  // Add this
    },
    token: string,
  ): Promise<{
    articles: Article[];
    page: number;
    page_size: number;
    total: number;
    has_more: boolean;
  }>;

  // Add new methods
  addTags(id: string, tags: string[], token: string): Promise<{ tags: string[] }>;
  setTags(id: string, tags: string[], token: string): Promise<{ tags: string[] }>;
  getTags(id: string, token: string): Promise<{ tags: string[] }>;
  removeTags(id: string, tags: string[], token: string): Promise<{ tags: string[] }>;
  getAllTags(token: string): Promise<{ tags: string[] }>;
}
```

#### Update `createApiClient()` implementation:

1. Update `getArticles()` to include tag query parameter:

```typescript
getArticles: (
  params: {
    page?: number;
    page_size?: number;
    favorite?: boolean;
    tag?: string;
  },
  token: string,
) => {
  const queryParams = new URLSearchParams();
  if (params.page !== undefined) queryParams.set("page", params.page.toString());
  if (params.page_size !== undefined) queryParams.set("page_size", params.page_size.toString());
  if (params.favorite !== undefined) queryParams.set("favorite", params.favorite.toString());
  if (params.tag !== undefined) queryParams.set("tag", params.tag);  // Add this

  const path = `/v1/articles${queryParams.toString() ? `?${queryParams.toString()}` : ""}`;
  return request<...>("GET", path, token);
},
```

2. Add tag management methods:

```typescript
addTags: (id: string, tags: string[], token: string) =>
  request<{ tags: string[] }>("POST", `/v1/articles/${id}/tags`, token, { tags }),

setTags: (id: string, tags: string[], token: string) =>
  request<{ tags: string[] }>("PUT", `/v1/articles/${id}/tags`, token, { tags }),

getTags: (id: string, token: string) =>
  request<{ tags: string[] }>("GET", `/v1/articles/${id}/tags`, token),

removeTags: (id: string, tags: string[], token: string) =>
  request<{ tags: string[] }>("DELETE", `/v1/articles/${id}/tags`, token, { tags }),

getAllTags: (token: string) =>
  request<{ tags: string[] }>("GET", "/v1/tags", token),
```

**Why:** Provides the API layer for tag operations needed by the webapp.

**Risks:** None - straightforward API additions matching backend endpoints.

---

## Phase 2: Webapp Server API Client (`frontend/webapp/src/lib/server/apiClient.ts`)

### Step 2.1: Update getArticles Parameters

Update function signature to accept `tag` parameter:

```typescript
export function getArticles(
  event: RequestEvent,
  params: { page?: number; page_size?: number; favorite?: boolean; tag?: string }
) {
  const client = createApiClient(event);
  return withSvelteKitError(() => client.getArticles(params, event.locals.auth ?? ''));
}
```

### Step 2.2: Add Tag Management Functions

Add new functions following the existing pattern:

```typescript
export function addTags(event: RequestEvent, id: string, tags: string[]) {
  const client = createApiClient(event);
  return withSvelteKitError(() => client.addTags(id, tags, event.locals.auth ?? ''));
}

export function setTags(event: RequestEvent, id: string, tags: string[]) {
  const client = createApiClient(event);
  return withSvelteKitError(() => client.setTags(id, tags, event.locals.auth ?? ''));
}

export function getTags(event: RequestEvent, id: string) {
  const client = createApiClient(event);
  return withSvelteKitError(() => client.getTags(id, event.locals.auth ?? ''));
}

export function removeTags(event: RequestEvent, id: string, tags: string[]) {
  const client = createApiClient(event);
  return withSvelteKitError(() => client.removeTags(id, tags, event.locals.auth ?? ''));
}

export function getAllTags(event: RequestEvent) {
  const client = createApiClient(event);
  return withSvelteKitError(() => client.getAllTags(event.locals.auth ?? ''));
}
```

**Why:** Wraps shared library API calls with SvelteKit-specific error handling.

**Risks:** None - follows existing patterns (e.g., `favoriteArticle`, `deleteArticle`).

---

## Phase 3: Reusable UI Components

### Step 3.1: Create TagsInput Component

**New file:** `frontend/webapp/src/lib/components/TagsInput.svelte`

**Purpose:** Reusable tag input component for adding/editing tags.

**Features:**
- Input field for comma-separated tag entry
- Real-time tag count display (X/10)
- Tag length validation (max 50 chars each)
- Visual feedback for validation errors
- Remove individual tags with "×" button
- Auto-normalize tags (trim, lowercase)

**Props:**

```typescript
interface Props {
  value: string[];              // Array of tags
  onChange: (tags: string[]) => void;  // Called when tags change
  disabled?: boolean;           // Disable input
  maxTags?: number;             // Default 10
  maxTagLength?: number;        // Default 50
  error?: string | null;        // Validation error message
}
```

**Example implementation structure:**

```svelte
<script lang="ts">
  let {
    value = [],
    onChange,
    disabled = false,
    maxTags = 10,
    maxTagLength = 50,
    error = null
  }: Props = $props();

  let input = $state('');

  function addTag(tag: string) {
    const normalized = tag.trim().toLowerCase();
    if (!normalized) return;

    if (value.includes(normalized)) {
      return; // Duplicate
    }

    if (value.length >= maxTags) {
      return; // Max reached
    }

    onChange([...value, normalized]);
    input = '';
  }

  function removeTag(tag: string) {
    onChange(value.filter(t => t !== tag));
  }

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === 'Enter' || event.key === ',') {
      event.preventDefault();
      addTag(input);
    }
  }
</script>

<div class="tags-input">
  <ul class="tags-list">
    {#each value as tag (tag)}
      <li class="tag">
        #{tag}
        <button type="button" onclick={() => removeTag(tag)} disabled={disabled}>×</button>
      </li>
    {/each}
  </ul>

  <input
    type="text"
    bind:value={input}
    placeholder={value.length === 0 ? 'Add tags...' : ''}
    disabled={disabled || value.length >= maxTags}
    onkeydown={handleKeydown}
  />

  <small class="tag-count">
    {value.length}/{maxTags}
    {#if error}
      <span class="error">{error}</span>
    {/if}
  </small>
</div>

<style>
  /* Pico.css friendly styles */
</style>
```

**Why:** Provides consistent tag editing UX across the application.

**Risks:**
- Need to ensure proper state synchronization.
- Edge cases: empty tags, duplicate tags (should be handled by backend but can pre-validate).

---

### Step 3.2: Create TagsEditor Component

**New file:** `frontend/webapp/src/lib/components/TagsEditor.svelte`

**Purpose:** Full tag management UI with add/remove/edit capabilities.

**Features:**
- Combines `TagsInput` with current tags display
- Show/hide edit mode toggle
- Auto-save or manual save option
- Loading states during save operations
- Error display

**Props:**

```typescript
interface Props {
  articleId: string;
  initialTags: string[];
  onSave: (tags: string[]) => Promise<void>;
  maxTags?: number;
  maxTagLength?: number;
}
```

**Example implementation structure:**

```svelte
<script lang="ts">
  import TagsInput from './TagsInput.svelte';

  let {
    articleId,
    initialTags,
    onSave,
    maxTags = 10,
    maxTagLength = 50
  }: Props = $props();

  let isEditing = $state(false);
  let tags = $state([...initialTags]);
  let isSaving = $state(false);
  let error = $state<string | null>(null);

  async function handleSave() {
    isSaving = true;
    error = null;

    try {
      await onSave(tags);
      isEditing = false;
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to save tags';
    } finally {
      isSaving = false;
    }
  }

  function handleCancel() {
    tags = [...initialTags];
    isEditing = false;
    error = null;
  }
</script>

<div class="tags-editor">
  {#if isEditing}
    <TagsInput
      bind:value={tags}
      disabled={isSaving}
      {maxTags}
      {maxTagLength}
      {error}
    />
    <div class="actions">
      <button type="button" onclick={handleSave} disabled={isSaving}>
        {isSaving ? 'Saving...' : 'Save'}
      </button>
      <button type="button" onclick={handleCancel} disabled={isSaving}>
        Cancel
      </button>
    </div>
  {:else}
    <div class="tags-display" onclick={() => isEditing = true}>
      {#if tags.length > 0}
        <ul>
          {#each tags as tag (tag)}
            <li>#{tag}</li>
          {/each}
        </ul>
      {:else}
        <span class="no-tags">Add tags...</span>
      {/if}
      <button type="button" class="edit-button" aria-label="Edit tags">✎</button>
    </div>
  {/if}
</div>

<style>
  /* Pico.css friendly styles */
</style>
```

**Why:** Provides a comprehensive tag editing experience for article detail pages.

**Risks:**
- Race conditions if user edits while another operation is in progress.
- Error handling and rollback on failed save.

---

### Step 3.3: Update Tags Component (Optional Enhancement)

**File:** `frontend/webapp/src/lib/components/Tags.svelte`

**Changes:** Add optional interactive features:

```typescript
let {
  tags,
  clickable = false,  // New prop
  onTagClick = (tag: string) => void 0  // Optional callback
}: { tags: Article['tags']; clickable?: boolean; onTagClick?: (tag: string) => void } = $props();
```

Update the tag display to be clickable when enabled:

```svelte
{#if tags && tags.length > 0}
  <ul>
    {#each tags as tag (tag)}
      <li>
        {#if clickable}
          <a href="/articles?tag={encodeURIComponent(tag)}" onclick={() => onTagClick(tag)}>
            #{tag}
          </a>
        {:else}
          <ins>#{tag}</ins>
        {/if}
      </li>
    {/each}
  </ul>
{/if}
```

**Why:** Improves UX by allowing users to click tags to filter articles.

**Risks:**
- Need to ensure existing usages (ArticleMetaItem, article detail page) still work.
- Link routing needs to handle current page context.

---

## Phase 4: Article Detail Page Enhancements

### Step 4.1: Add Server Actions for Tag Operations

**File:** `frontend/webapp/src/routes/articles/[id]/+page.server.ts`

Add tag management actions to the `actions` object:

```typescript
export const actions = {
  // ... existing actions (favorite, send, delete)

  setTags: async ({ locals, fetch, params, request, getClientAddress }) => {
    const data = await request.formData();
    const tagsStr = data.get('tags');

    if (typeof tagsStr !== 'string') {
      return fail(400, { message: 'Invalid tags format' });
    }

    try {
      const tags = JSON.parse(tagsStr);
      if (!Array.isArray(tags) || !tags.every((t) => typeof t === 'string')) {
        return fail(400, { message: 'Tags must be an array of strings' });
      }

      const result = await withActionFail(() =>
        setTags({ locals, fetch, request, getClientAddress } as RequestEvent, params.id, tags)
      );
      return result;
    } catch (e) {
      return fail(400, { message: 'Invalid JSON format' });
    }
  }
} satisfies Actions;
```

**Why:** Enables form-based tag updates with SvelteKit progressive enhancement.

**Risks:**
- JSON parsing errors - need proper error handling.
- Validation should match backend rules (max 10 tags, 50 chars each).

---

### Step 4.2: Update Article Detail Page UI

**File:** `frontend/webapp/src/routes/articles/[id]/+page.svelte`

**Changes:**

1. Import the TagsEditor component:

```typescript
import TagsEditor from '$lib/components/TagsEditor.svelte';
```

2. Add state management for tag editing and form reference:

```typescript
let tagsForm = $state<HTMLFormElement>();
let tagsError = $state<string | null>(null);
```

3. Replace the read-only `Tags` component with `TagsEditor`:

```svelte
<!-- Before -->
<Tags tags={data.tags} />

<!-- After -->
<TagsEditor
  articleId={data.id}
  initialTags={data.tags || []}
  onSave={async (tags) => {
    tagsError = null;
    // Submit form with tags data
    if (tagsForm) {
      const formData = new FormData();
      formData.append('tags', JSON.stringify(tags));
      await new Promise((resolve) => {
        const form = new HTMLFormElement();
        const submitEvent = new SubmitEvent('submit', { submitter: form });
        form.requestSubmit(submitEvent);
      });
    }
  }}
/>
```

4. Add a hidden form for the server action:

```svelte
<form
  bind:this={tagsForm}
  method="POST"
  action="/articles/{data.id}?/setTags"
  use:enhance={() => {
    return async ({ update }: { update: () => Promise<void> }) => {
      await update();
      // Refresh article data
    };
  }}
></form>
```

5. Display error if any:

```svelte
{#if tagsError}
  <p style="color: red" role="alert">{tagsError}</p>
{/if}
```

**Why:** Allows users to manage tags directly on the article detail page.

**Risks:**
- Need to invalidate data after tag update (article list may need refresh).
- Integration with existing ArticleControls component.

---

## Phase 5: Article List Page - Tag Filtering

### Step 5.1: Update Server Load Function

**File:** `frontend/webapp/src/routes/articles/+page.server.ts`

Add tag parameter parsing and pass to API:

```typescript
export const load: PageServerLoad = async ({ locals, fetch, url, request, getClientAddress }) => {
  const pageParam = url.searchParams.get('page');
  const pageSizeParam = url.searchParams.get('page_size');
  const favoritesParam = url.searchParams.get('favorites');
  const tagParam = url.searchParams.get('tag');  // Add this

  const page = pageParam ? parseInt(pageParam, 10) : 1;
  const pageSize = pageSizeParam ? parseInt(pageSizeParam, 10) : 10;
  const favorite = favoritesParam === 'true' ? true : undefined;
  const tag = tagParam || undefined;  // Add this

  const data = await getArticles({ locals, fetch, request, getClientAddress } as RequestEvent, {
    page,
    page_size: pageSize,
    favorite,
    tag  // Add this
  });

  return { ...data, user: locals.user };
};
```

**Why:** Enables filtering articles by tag.

**Risks:**
- Need to handle pagination with tag filter.
- Should work alongside existing `favorite` filter.

---

### Step 5.2: Update Article List Page UI

**File:** `frontend/webapp/src/routes/articles/+page.svelte`

**Changes:**

1. Add state for current tag filter and all available tags:

```typescript
let currentTag = $derived(data.url?.searchParams.get('tag') || null);
let allTags = $state<string[]>([]);
let showTagFilter = $state(false);

// Load all tags on mount
onMount(async () => {
  // Could fetch all tags here if needed for autocomplete
});
```

2. Add tag filter UI component:

```svelte
<div class="tag-filter">
  <button
    type="button"
    onclick={() => showTagFilter = !showTagFilter}
    aria-expanded={showTagFilter}
  >
    {#if currentTag}
      Filtered by: <strong>#{currentTag}</strong>
    {:else}
      Filter by tag
    {/if}
  </button>

  {#if showTagFilter}
    <input
      type="text"
      placeholder="Search tags..."
      bind:value={tagSearchInput}
      oninput={handleTagSearch}
    />

    {#if filteredTags.length > 0}
      <ul class="tag-suggestions">
        {#each filteredTags as tag (tag)}
          <li>
            <a href="/articles?tag={encodeURIComponent(tag)}">#{tag}</a>
          </li>
        {/each}
      </ul>
    {:else if tagSearchInput}
      <p>No tags found</p>
    {/if}
  {/if}

  {#if currentTag}
    <a href="/articles" class="clear-filter">Clear filter</a>
  {/if}
</div>
```

3. Update ArticleMetaItem to show tags as clickable links:

```svelte
<!-- In the article item template -->
<Tags tags={article.tags} clickable={true} />
```

4. Update URL management to preserve other filters:

```typescript
function navigateWithTag(tag: string) {
  const url = new URL(window.location.href);
  url.searchParams.set('tag', tag);
  window.location.href = url.toString();
}

function clearTagFilter() {
  const url = new URL(window.location.href);
  url.searchParams.delete('tag');
  window.location.href = url.toString();
}
```

**Why:** Provides easy navigation and discovery via tags.

**Risks:**
- URL management complexity (preserving page number, favorite filter).
- UX: what happens if no articles have a tag?
- Performance: loading all tags for suggestions.

---

## Phase 6: Tags Management Page (Optional Enhancement)

### Step 6.1: Create Tags Route

**New file:** `frontend/webapp/src/routes/tags/+page.server.ts`

```typescript
import { getAllTags, getArticles } from '$lib/server/apiClient';
import type { PageServerLoad, RequestEvent } from './$types';

export const load: PageServerLoad = async ({ locals, fetch, request, getClientAddress }) => {
  const allTags = await getAllTags({ locals, fetch, request, getClientAddress } as RequestEvent);

  // Optionally get article counts for each tag
  const tagsWithCounts = await Promise.all(
    allTags.tags.map(async (tag) => {
      const result = await getArticles(
        { locals, fetch, request, getClientAddress } as RequestEvent,
        { tag, page: 1, page_size: 1 }  // Just get total count
      );
      return { tag, count: result.total };
    })
  );

  return { tags: tagsWithCounts.sort((a, b) => b.count - a.count), user: locals.user };
};
```

**New file:** `frontend/webapp/src/routes/tags/+page.svelte`

```svelte
<script lang="ts">
  import { goto } from '$app/navigation';

  let { data }: { data: { tags: { tag: string; count: number }[] } } = $props();

  function filterByTag(tag: string) {
    goto(`/articles?tag=${encodeURIComponent(tag)}`);
  }
</script>

<section>
  <hgroup>
    <h1>Tags</h1>
    <p>Browse and manage your article tags</p>
  </hgroup>

  {#if data.tags.length === 0}
    <p>No tags yet. Add tags to your articles to see them here.</p>
  {:else}
    <ul class="tags-cloud">
      {#each data.tags as { tag, count } (tag)}
        <li>
          <button type="button" onclick={() => filterByTag(tag)}>
            #{tag} <span class="count">{count}</span>
          </button>
        </li>
      {/each}
    </ul>
  {/if}
</section>

<style>
  .tags-cloud {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    padding: 0;
  }

  .tags-cloud li {
    list-style: none;
  }

  .tags-cloud button {
    display: inline-flex;
    align-items: center;
    gap: 0.25rem;
  }

  .count {
    font-size: 0.8em;
    opacity: 0.7;
  }
</style>
```

**Purpose:** Dedicated page to view and manage all tags.

**Features:**
- List all tags with article counts
- Filter articles by clicking a tag
- Sort by article count or alphabetically

**Why:** Centralized tag management for power users.

**Risks:**
- Backend doesn't support rename/delete operations yet (only add/remove from articles).
- Article count queries may be expensive (multiple API calls).

---

### Step 6.2: Add Navigation Link

**File:** `frontend/webapp/src/lib/components/Nav.svelte`

Add "Tags" link to navigation:

```typescript
const links = [
  { href: '/new', label: 'New' },
  { href: '/articles', label: 'Articles' },
  { href: '/articles?favorites=true', label: 'Favorites' },
  { href: '/tags', label: 'Tags' },  // Add this
  { href: '/account', label: 'Account' }
] as const;
```

**Why:** Makes tags page discoverable.

**Risks:** None - simple addition.

---

## Phase 7: Testing

### Step 7.1: Unit Tests

**New test file:** `frontend/webapp/src/lib/components/TagsInput.svelte.spec.ts`

```typescript
import { describe, expect, it } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/svelte/svelte5';
import TagsInput from './TagsInput.svelte';

describe('TagsInput', () => {
  it('should render initial tags', async () => {
    const tags = ['tech', 'reading'];
    const { container } = render(TagsInput, { value: tags, onChange: () => {} });

    expect(screen.getByText('#tech')).toBeInTheDocument();
    expect(screen.getByText('#reading')).toBeInTheDocument();
  });

  it('should call onChange when a tag is removed', async () => {
    const onChange = vi.fn();
    const tags = ['tech', 'reading'];

    render(TagsInput, { value: tags, onChange });

    const removeButton = screen.getByLabelText(/remove reading/i);
    await fireEvent.click(removeButton);

    expect(onChange).toHaveBeenCalledWith(['tech']);
  });

  it('should not add duplicate tags', async () => {
    const onChange = vi.fn();
    const tags = ['tech'];

    render(TagsInput, { value: tags, onChange });

    const input = screen.getByPlaceholderText(/add tags/i);
    await fireEvent.update(input, 'tech');
    await fireEvent.keyDown(input, { key: 'Enter' });

    expect(onChange).not.toHaveBeenCalled();
  });

  it('should not add more than max tags', async () => {
    const onChange = vi.fn();
    const tags = Array(10).fill('tag');

    render(TagsInput, { value: tags, onChange, maxTags: 10 });

    const input = screen.getByPlaceholderText(/add tags/i);
    await fireEvent.update(input, 'newtag');
    await fireEvent.keyDown(input, { key: 'Enter' });

    expect(onChange).not.toHaveBeenCalled();
  });

  it('should normalize tags to lowercase', async () => {
    const onChange = vi.fn();

    render(TagsInput, { value: [], onChange });

    const input = screen.getByPlaceholderText(/add tags/i);
    await fireEvent.update(input, 'Tech Tag');
    await fireEvent.keyDown(input, { key: 'Enter' });

    expect(onChange).toHaveBeenCalledWith(['tech tag']);
  });
});
```

**New test file:** `frontend/webapp/src/lib/components/TagsEditor.svelte.spec.ts`

```typescript
import { describe, expect, it, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/svelte/svelte5';
import TagsEditor from './TagsEditor.svelte';

describe('TagsEditor', () => {
  it('should display tags in view mode', () => {
    const tags = ['tech', 'reading'];

    render(TagsEditor, {
      articleId: '123',
      initialTags: tags,
      onSave: vi.fn()
    });

    expect(screen.getByText('#tech')).toBeInTheDocument();
    expect(screen.getByText('#reading')).toBeInTheDocument();
  });

  it('should enter edit mode when clicking on tags', async () => {
    render(TagsEditor, {
      articleId: '123',
      initialTags: ['tech'],
      onSave: vi.fn()
    });

    const tagsDisplay = screen.getByRole('list');
    await fireEvent.click(tagsDisplay);

    // Should show edit mode controls
    expect(screen.getByRole('button', { name: /save/i })).toBeInTheDocument();
  });

  it('should call onSave when saving tags', async () => {
    const onSave = vi.fn().mockResolvedValue(undefined);

    render(TagsEditor, {
      articleId: '123',
      initialTags: ['tech'],
      onSave
    });

    // Enter edit mode
    const tagsDisplay = screen.getByRole('list');
    await fireEvent.click(tagsDisplay);

    // Save
    const saveButton = screen.getByRole('button', { name: /save/i });
    await fireEvent.click(saveButton);

    await waitFor(() => {
      expect(onSave).toHaveBeenCalledWith(['tech']);
    });
  });

  it('should cancel edits when clicking cancel', async () => {
    const onSave = vi.fn();
    const initialTags = ['tech', 'reading'];

    render(TagsEditor, {
      articleId: '123',
      initialTags,
      onSave
    });

    // Enter edit mode
    const tagsDisplay = screen.getByRole('list');
    await fireEvent.click(tagsDisplay);

    // Modify tags
    const input = screen.getByPlaceholderText(/add tags/i);
    await fireEvent.update(input, 'newtag');
    await fireEvent.keyDown(input, { key: 'Enter' });

    // Cancel
    const cancelButton = screen.getByRole('button', { name: /cancel/i });
    await fireEvent.click(cancelButton);

    // Should revert to original tags
    expect(screen.getByText('#tech')).toBeInTheDocument();
    expect(screen.getByText('#reading')).toBeInTheDocument();
    expect(screen.queryByText('#newtag')).not.toBeInTheDocument();
  });
});
```

**Extend:** `frontend/shared/src/lib/apiClient.spec.ts`

Add tests for new tag methods:

```typescript
describe('tag methods', () => {
  it('should add tags to article', async () => {
    const mockFetch = vi.fn(() =>
      Promise.resolve({
        ok: true,
        json: () => Promise.resolve({ tags: ['tech', 'reading'] })
      })
    );

    const client = createApiClient({ baseUrl: 'http://test.com', fetch: mockFetch });
    const result = await client.addTags('article123', ['reading'], 'token123');

    expect(result).toEqual({ tags: ['tech', 'reading'] });
    expect(mockFetch).toHaveBeenCalledWith(
      'http://test.com/v1/articles/article123/tags',
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ tags: ['reading'] })
      })
    );
  });

  it('should set tags for article', async () => {
    const mockFetch = vi.fn(() =>
      Promise.resolve({
        ok: true,
        json: () => Promise.resolve({ tags: ['tech'] })
      })
    );

    const client = createApiClient({ baseUrl: 'http://test.com', fetch: mockFetch });
    const result = await client.setTags('article123', ['tech'], 'token123');

    expect(result).toEqual({ tags: ['tech'] });
    expect(mockFetch).toHaveBeenCalledWith(
      'http://test.com/v1/articles/article123/tags',
      expect.objectContaining({
        method: 'PUT',
        body: JSON.stringify({ tags: ['tech'] })
      })
    );
  });

  it('should remove tags from article', async () => {
    const mockFetch = vi.fn(() =>
      Promise.resolve({
        ok: true,
        json: () => Promise.resolve({ tags: ['tech'] })
      })
    );

    const client = createApiClient({ baseUrl: 'http://test.com', fetch: mockFetch });
    const result = await client.removeTags('article123', ['reading'], 'token123');

    expect(result).toEqual({ tags: ['tech'] });
    expect(mockFetch).toHaveBeenCalledWith(
      'http://test.com/v1/articles/article123/tags',
      expect.objectContaining({
        method: 'DELETE',
        body: JSON.stringify({ tags: ['reading'] })
      })
    );
  });

  it('should get all tags for account', async () => {
    const mockFetch = vi.fn(() =>
      Promise.resolve({
        ok: true,
        json: () => Promise.resolve({ tags: ['tech', 'reading', 'tutorial'] })
      })
    );

    const client = createApiClient({ baseUrl: 'http://test.com', fetch: mockFetch });
    const result = await client.getAllTags('token123');

    expect(result).toEqual({ tags: ['tech', 'reading', 'tutorial'] });
    expect(mockFetch).toHaveBeenCalledWith(
      'http://test.com/v1/tags',
      expect.objectContaining({
        method: 'GET'
      })
    );
  });
});
```

**Extend:** `frontend/webapp/src/routes/articles/[id]/page.server.test.ts`

Add tests for tag actions:

```typescript
describe('setTags action', () => {
  it('should set tags successfully', async () => {
    const mockSetTags = vi.fn().mockResolvedValue({ tags: ['tech', 'reading'] });
    vi.mock('$lib/server/apiClient', () => ({
      ...await import('$lib/server/apiClient'),
      setTags: mockSetTags,
      withActionFail: (fn) => fn()
    }));

    const { actions } = await import('./+page.server');

    const formData = new FormData();
    formData.append('tags', JSON.stringify(['tech', 'reading']));

    const mockRequest = new Request('http://localhost/articles/123', {
      method: 'POST',
      body: formData
    });

    const result = await actions.setTags({
      ...mockRequestEvent,
      request: mockRequest,
      params: { id: '123' }
    } as any);

    expect(mockSetTags).toHaveBeenCalledWith(
      expect.any(Object),
      '123',
      ['tech', 'reading']
    );
  });

  it('should return error for invalid JSON', async () => {
    const { actions } = await import('./+page.server');

    const formData = new FormData();
    formData.append('tags', 'invalid json');

    const mockRequest = new Request('http://localhost/articles/123', {
      method: 'POST',
      body: formData
    });

    const result = await actions.setTags({
      ...mockRequestEvent,
      request: mockRequest,
      params: { id: '123' }
    } as any);

    expect(result).toHaveProperty('status', 400);
    expect(result?.data).toHaveProperty('message', 'Invalid JSON format');
  });
});
```

**Test coverage goals:**
- Tag validation logic (max tags, max length)
- Tag normalization (trimming, deduplication)
- API client method behavior
- Server actions for tag operations
- Error handling (invalid tags, network errors)

---

### Step 7.2: Integration Tests (Optional)

**E2E tests using Playwright:**

**New file:** `frontend/webapp/tests/tags.spec.ts`

```typescript
import { expect, test } from '@playwright/test';

test.describe('Tag Management', () => {
  test('should add tags to new article', async ({ page }) => {
    await page.goto('/new');

    await page.fill('input[name="url"]', 'https://example.com/article');
    await page.fill('input[name="tags"]', 'tech, reading, tutorial');
    await page.click('button[type="submit"]');

    // Should navigate to articles list
    await expect(page).toHaveURL('/articles');
  });

  test('should edit tags on article detail page', async ({ page }) => {
    await page.goto('/articles');

    // Navigate to first article
    await page.click('article a');

    // Click edit tags
    await page.click('.tags-editor .tags-display');

    // Remove a tag
    await page.click('.tag button[aria-label*="remove"]');

    // Save
    await page.click('button:text("Save")');

    // Verify tag removed
    await expect(page.locator('.tags-editor')).toContainText('#tech');
    await expect(page.locator('.tags-editor')).not.toContainText('#reading');
  });

  test('should filter articles by tag', async ({ page }) => {
    await page.goto('/articles');

    // Click a tag on an article
    await page.click('a[href*="tag="]');

    // Verify URL has tag parameter
    await expect(page).toHaveURL(/tag=/);

    // Verify filtered results
    const currentUrl = page.url();
    const tag = new URL(currentUrl).searchParams.get('tag');
    expect(tag).toBeTruthy();
  });

  test('should clear tag filter', async ({ page }) => {
    await page.goto('/articles?tag=tech');

    // Click clear filter
    await page.click('.clear-filter');

    // Verify tag parameter removed
    await expect(page).toHaveURL('/articles');
    await expect(page.locator('.tag-filter')).not.toContainText('Filtered by');
  });
});
```

**Why:** Ensures full user flows work end-to-end.

**Risks:**
- Need to ensure test data cleanup.
- May need to mock or configure backend test server.

---

## Summary of Files to Modify

### New Files:
1. `frontend/shared/src/types/schemas.ts` - Export tag types
2. `frontend/webapp/src/lib/components/TagsInput.svelte` - Tag input component
3. `frontend/webapp/src/lib/components/TagsEditor.svelte` - Tag editor component
4. `frontend/webapp/src/routes/tags/+page.server.ts` - Tags page server load
5. `frontend/webapp/src/routes/tags/+page.svelte` - Tags page UI
6. `frontend/webapp/src/lib/components/TagsInput.svelte.spec.ts` - Tests
7. `frontend/webapp/src/lib/components/TagsEditor.svelte.spec.ts` - Tests
8. `frontend/webapp/tests/tags.spec.ts` - E2E tests (optional)

### Modified Files:
1. `frontend/shared/src/lib/apiClient.ts` - Add tag methods and update getArticles
2. `frontend/webapp/src/lib/server/apiClient.ts` - Add tag wrappers
3. `frontend/webapp/src/lib/components/Tags.svelte` - Add clickable links (optional)
4. `frontend/webapp/src/routes/articles/[id]/+page.server.ts` - Add tag actions
5. `frontend/webapp/src/routes/articles/[id]/+page.svelte` - Add tag editor
6. `frontend/webapp/src/routes/articles/+page.server.ts` - Add tag filtering
7. `frontend/webapp/src/routes/articles/+page.svelte` - Add tag filter UI
8. `frontend/webapp/src/lib/components/Nav.svelte` - Add Tags link
9. `frontend/shared/src/lib/apiClient.spec.ts` - Extend tests
10. `frontend/webapp/src/routes/articles/[id]/page.server.test.ts` - Extend tests

---

## Implementation Order

**Recommended order for minimal disruption:**

1. **Phase 1**: Shared library (foundation)
   - Step 1.1: Export tag types
   - Step 1.2: Update ApiClient

2. **Phase 2**: Webapp server API client
   - Step 2.1: Update getArticles
   - Step 2.2: Add tag management functions

3. **Phase 3**: UI components (independent)
   - Step 3.1: Create TagsInput component
   - Step 3.2: Create TagsEditor component
   - Step 3.3: Update Tags component (optional)

4. **Phase 4**: Article detail page (core feature)
   - Step 4.1: Add server actions
   - Step 4.2: Update article detail page UI

5. **Phase 5**: Article list filtering (core feature)
   - Step 5.1: Update server load function
   - Step 5.2: Update article list page UI

6. **Phase 6**: Tags page (optional enhancement)
   - Step 6.1: Create tags route
   - Step 6.2: Add navigation link

7. **Phase 7**: Testing (throughout development)
   - Step 7.1: Unit tests
   - Step 7.2: Integration tests (optional)

---

## Key Considerations & Edge Cases

### 1. Tag Normalization
**Issue:** Backend normalizes tags (lowercase, trimmed, deduplicated).

**Solution:** Frontend should pre-validate for better UX, but trust backend for final normalization.

---

### 2. Error Handling
**Issues:**
- API errors (400, 401, 404, 500)
- Network errors
- Validation errors

**Solution:**
- Display user-friendly error messages
- Use toast notifications or inline error messages
- Provide "Try again" option for network errors

---

### 3. Loading States
**Issue:** Tag operations take time and need visual feedback.

**Solution:**
- Show loading indicators (spinner or progress bar)
- Disable inputs/buttons during operations
- Use optimistic updates where appropriate

---

### 4. Optimistic Updates
**Issue:** Delay between action and server response can feel sluggish.

**Solution:**
- Update UI immediately when possible
- Rollback on error
- Especially useful for add/remove operations

---

### 5. URL State Management
**Issue:** Tag filter state should be shareable/bookmarkable.

**Solution:**
- Store tag filter in URL query params: `?tag=tech`
- Preserve other filters (page, favorite)
- Use SvelteKit's URL handling for navigation

---

### 6. Pagination with Tag Filtering
**Issue:** Pagination needs to work with tag filters.

**Solution:**
- Pass tag parameter to getArticles
- Preserve tag in pagination links
- Reset page to 1 when tag changes

---

### 7. Concurrent Edits
**Issue:** User edits tags while another operation is in progress.

**Solution:**
- Disable edit mode during operations
- Show loading state
- Queue or reject concurrent operations

---

### 8. Accessibility
**Requirements:**
- Proper ARIA labels for tag controls
- Keyboard navigation for tag editing
- Screen reader announcements for tag operations

**Solution:**
- Use semantic HTML (buttons, links)
- Add `aria-label` to icon-only buttons
- Announce changes with `aria-live` regions

---

### 9. Mobile Responsiveness
**Issue:** Tag input and filter UI should work well on mobile.

**Solution:**
- Use flexible layouts
- Make tags clickable with adequate touch targets
- Consider collapsible tag filter on mobile

---

### 10. Data Refresh
**Issue:** After tag updates, related data needs to be refreshed.

**Solution:**
- Use SvelteKit's `invalidate()` function
- Invalidate article list after tag changes
- Consider invalidating tag list on the tags page

---

### 11. Empty States
**Issues:**
- What to show when article has no tags?
- What if no articles have a particular tag?
- What if user has no tags at all?

**Solution:**
- Show "Add tags..." or similar message
- Show empty state when filtering returns no results
- Show helpful message on tags page

---

### 12. Tag Validation Feedback
**Issue:** Users need to know when their tag input is invalid.

**Solution:**
- Real-time validation feedback
- Show error messages below input
- Highlight invalid input
- Prevent submission with invalid tags

---

## Validation Rules to Enforce (Matching Backend)

```typescript
const MAX_TAGS = 10;
const MAX_TAG_LENGTH = 50;

function validateTags(tags: string[]): string | null {
  if (tags.length > MAX_TAGS) {
    return `Maximum ${MAX_TAGS} tags allowed per article`;
  }

  for (const tag of tags) {
    if (tag.length === 0) {
      return 'Tags cannot be empty';
    }
    if (tag.length > MAX_TAG_LENGTH) {
      return `Tag "${tag}" exceeds maximum length of ${MAX_TAG_LENGTH} characters`;
    }
  }

  return null; // Valid
}

function normalizeTags(tags: string[]): string[] {
  return tags
    .map(tag => tag.trim().toLowerCase())
    .filter(tag => tag.length > 0)
    .filter((tag, index, self) => self.indexOf(tag) === index); // Deduplicate
}
```

---

## Future Enhancements (Beyond Initial Implementation)

1. **Tag Renaming**: Allow users to rename tags across all articles
2. **Tag Deletion**: Allow bulk deletion of tags from all articles
3. **Tag Merging**: Merge similar tags (e.g., "js" and "javascript")
4. **Tag Autocomplete**: Suggest existing tags based on typing
5. **Tag Colors/Icons**: Visual differentiation for tags
6. **Tag Groups/Categories**: Organize tags into groups
7. **Tag Statistics**: Show most-used tags, tag trends
8. **Bulk Tag Operations**: Apply tags to multiple articles at once
9. **Tag Export/Import**: Backup and restore tag configuration
10. **Tag Search**: Full-text search across tags

---

## Testing Checklist

- [ ] Tag input component unit tests
- [ ] Tag editor component unit tests
- [ ] API client tag methods tests
- [ ] Server action tests for tag operations
- [ ] Tag filtering in article list
- [ ] Tag editing on article detail page
- [ ] Error handling for invalid tags
- [ ] Error handling for network failures
- [ ] Accessibility testing (keyboard nav, screen reader)
- [ ] Mobile responsiveness testing
- [ ] E2E tests for complete user flows

---

## Definition of Done

A feature is complete when:

1. **Code Implementation**: All phases are implemented and code is written
2. **Tests**: Unit tests are written and passing
3. **Linting**: `just check` passes without errors
4. **Manual Testing**: Feature is tested manually in the browser
5. **Documentation**: This plan is updated with any implementation notes
6. **No Regressions**: Existing functionality remains intact

---

## Questions & Decisions

### Q1: Should we use optimistic updates for tag operations?

**Options:**
- **Yes**: Update UI immediately, rollback on error (better UX)
- **No**: Wait for server response before updating (simpler, less buggy)

**Recommendation**: Start with no optimistic updates for simplicity, add later if needed.

---

### Q2: Should tags on the article list be clickable by default?

**Options:**
- **Yes**: All tags are clickable links to filter by tag
- **No**: Make tags clickable only in specific contexts (e.g., detail page)

**Recommendation**: Make tags clickable on list view but provide a prop to disable.

---

### Q3: Should we implement the tags page in the initial implementation?

**Options:**
- **Yes**: Include dedicated tags page for comprehensive management
- **No**: Focus on list filtering first, add tags page later

**Recommendation**: Exclude from initial implementation, add as Phase 6.

---

### Q4: How should we handle tag suggestions/autocomplete?

**Options:**
- Fetch all tags on page load (simple, potential performance issue)
- Lazy load tags when user starts typing
- No autocomplete, manual tag entry only

**Recommendation**: Start with manual entry only, add autocomplete later if needed.

---

## OpenAPI Schema References

The following schemas from the backend OpenAPI spec are relevant:

**TagsRequest:**
```yaml
TagsRequest:
  type: object
  required:
    - tags
  properties:
    tags:
      type: array
      items:
        type: string
        maxLength: 50
      description: Tags to manage (1-10 tags, max 50 characters each)
      minItems: 1
      maxItems: 10
```

**TagsResponse:**
```yaml
TagsResponse:
  type: object
  required:
    - tags
  properties:
    tags:
      type: array
      items:
        type: string
      description: List of tags
```

**Article.tags:**
```yaml
tags:
  type: array
  items:
    type: string
  description: Tags associated with the article
```
