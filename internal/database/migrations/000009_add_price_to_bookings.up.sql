ALTER TABLE bookings ADD COLUMN price DECIMAL(10,2) NOT NULL DEFAULT 0;
UPDATE bookings SET price = gross_amount;
