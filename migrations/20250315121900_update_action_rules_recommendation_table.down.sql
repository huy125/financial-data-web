ALTER TABLE recommendation 
DROP CONSTRAINT recommendation_action_check, 
ADD CONSTRAINT recommendation_action_check 
CHECK (action IN ('Buy', 'Hold', 'Sell'));