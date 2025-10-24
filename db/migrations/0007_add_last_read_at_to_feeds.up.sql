-- Add last_read_at column to feeds table
ALTER TABLE feeds 
ADD COLUMN last_read_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();