-- Seed a sample public page for previewing theme + blocks.
WITH inserted_page AS (
    INSERT INTO pages (title, slug, parent_id, layout, is_published, metadata, path, sort_order)
    VALUES ('Sample Page', 'sample', NULL, 'default', true, '{}'::jsonb, 'sample', 0)
    ON CONFLICT (path) DO NOTHING
    RETURNING id
),
page AS (
    SELECT id FROM inserted_page
    UNION ALL
    SELECT id FROM pages WHERE path = 'sample'
),
seed_target AS (
    SELECT id FROM page
    WHERE NOT EXISTS (SELECT 1 FROM page_blocks WHERE page_id = page.id)
)
INSERT INTO page_blocks (page_id, block_type, content, sort_order)
SELECT id, 'hero', '{"headline":"Castogo: your podcast, your rules","subheadline":"Self-host your show with a clean workflow, beautiful public pages, and RSS that just works.","cta_text":"Explore episodes","cta_url":"/episodes","background_image":""}'::jsonb, 0
FROM seed_target
UNION ALL
SELECT id, 'features', '{"section_title":"Publish with confidence","section_description":"The building blocks you need for a polished podcast site.","items":[{"icon":"mic","title":"Episode workflow","description":"Draft, schedule, publish, and archive from one dashboard."},{"icon":"rss","title":"RSS done right","description":"Compliant feeds for Apple, Spotify, and Google."},{"icon":"layers","title":"Page builder","description":"Compose pages with reusable blocks and clean layouts."},{"icon":"shield","title":"Self-hosted","description":"Own your data and tune the experience to your brand."}]}'::jsonb, 1
FROM seed_target
UNION ALL
SELECT id, 'episodes_showcase', '{"section_title":"Latest episodes","section_description":"Fresh conversations and deep dives from the Castogo demo feed.","max_episodes":6,"display_mode":"grid"}'::jsonb, 2
FROM seed_target
UNION ALL
SELECT id, 'testimonials', '{"section_title":"Built for indie creators","section_description":"Simple tools that keep you focused on your show.","items":[{"quote":"We moved off a hosted platform and never looked back. Everything feels faster and clearer.","author":"Maya Liu","role":"Producer, The Signal","avatar_url":""},{"quote":"The blocks let us ship a landing page in an afternoon. It looks great on mobile.","author":"Jordan Patel","role":"Host, Night Shift","avatar_url":""},{"quote":"RSS just works. That alone saved us hours every release.","author":"Cam Torres","role":"Co-host, Field Notes","avatar_url":""}]}'::jsonb, 3
FROM seed_target
UNION ALL
SELECT id, 'prose', '{"body":"## About this sample page\n\nThis page is seeded from the database so you can preview public theming and blocks.\n\n- Hero, features, episodes, testimonials, and a CTA\n- Uses the **theme-public** token set\n- Fully editable from the admin Pages UI\n\n> Tip: create your own page in `/admin/pages` and publish it."}'::jsonb, 4
FROM seed_target
UNION ALL
SELECT id, 'cta', '{"headline":"Ready to build your public site?","description":"Jump into the admin and customize this page, or create a new one in minutes.","button_text":"Open admin","button_url":"/login"}'::jsonb, 5
FROM seed_target;
