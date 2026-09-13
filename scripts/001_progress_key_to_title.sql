--
--   sudo systemctl stop manga
--   sqlite3 /opt/manga/manga.db < 001_progress_key_to_title.sql
--   sudo systemctl start manga
--

ALTER TABLE reading_progress RENAME COLUMN mangadex_id TO title;

DELETE FROM reading_progress
WHERE title = ''
   OR title GLOB '[0-9a-f]*-[0-9a-f]*-[0-9a-f]*-[0-9a-f]*-[0-9a-f]*';
