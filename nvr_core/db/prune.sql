
-- Find the 100 oldest segments, delete their rows, and return their file paths in one atomic operation.
DELETE FROM segments 
WHERE id IN (
    SELECT id FROM segments ORDER BY start_time ASC LIMIT 100
)
RETURNING file_path;