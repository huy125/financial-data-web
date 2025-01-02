CREATE TABLE recommendation (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    analysis_id UUID NOT NULL REFERENCES analysis(id) ON DELETE CASCADE,
    action VARCHAR(50) NOT NULL CHECK (action IN ('Buy', 'Hold', 'Sell')),
    confidence_level FLOAT NOT NULL CHECK (confidence_level >= 0 AND confidence_level <= 100),
    reason TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
