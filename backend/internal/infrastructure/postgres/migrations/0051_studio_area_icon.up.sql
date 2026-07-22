ALTER TABLE studio_areas RENAME COLUMN emoji TO icon;
UPDATE studio_areas SET icon = CASE icon WHEN '☕' THEN 'coffee' WHEN '💼' THEN 'briefcase' WHEN '✈️' THEN 'plane' ELSE icon END;
