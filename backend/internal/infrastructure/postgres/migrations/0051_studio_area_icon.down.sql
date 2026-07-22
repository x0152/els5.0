UPDATE studio_areas SET icon = CASE icon WHEN 'coffee' THEN '☕' WHEN 'briefcase' THEN '💼' WHEN 'plane' THEN '✈️' ELSE icon END;
ALTER TABLE studio_areas RENAME COLUMN icon TO emoji;
