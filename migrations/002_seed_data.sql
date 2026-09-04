INSERT INTO shops (id, name, city, cuisine, delivery_time_min, is_active)
VALUES
    (1, 'Chai Corner', 'Moscow', 'Indian', 25, TRUE),
    (2, 'PastaLab', 'Moscow', 'Italian', 30, TRUE),
    (3, 'Sushi Metro', 'Moscow', 'Japanese', 22, TRUE)
ON CONFLICT (id) DO NOTHING;

INSERT INTO dishes (id, shop_id, name, description, category, price_cents, preparation_time_min, is_available)
VALUES
    (1, 1, 'Masala Chai', 'Black tea with cardamom and milk.', 'drinks', 250, 5, TRUE),
    (2, 1, 'Chicken Biryani', 'Fragrant rice with chicken and herbs.', 'main', 420, 20, TRUE),
    (3, 1, 'Paneer Wrap', 'Grilled paneer with salad and house sauce.', 'main', 390, 15, TRUE),
    (4, 2, 'Margherita Pizza', 'Tomato sauce, mozzarella and basil.', 'pizza', 540, 20, TRUE),
    (5, 2, 'Spaghetti Carbonara', 'Creamy pasta with bacon and parmesan.', 'pasta', 620, 20, TRUE),
    (6, 2, 'Caesar Salad', 'Romaine, chicken, parmesan and dressing.', 'salad', 430, 10, TRUE),
    (7, 3, 'Salmon Roll Set', 'Classic salmon sushi set with miso soup.', 'rolls', 760, 15, TRUE),
    (8, 3, 'Vegan Maki Box', 'Fresh vegetables and avocado in rice rolls.', 'maki', 490, 15, TRUE),
    (9, 3, 'Edamame Bowl', 'Warm edamame with sesame and chili.', 'snacks', 310, 8, TRUE)
ON CONFLICT (id) DO NOTHING;

INSERT INTO ingredients (id, name)
VALUES
    (1, 'Sugar'), (2, 'Milk'), (3, 'Cilantro'), (4, 'Chicken'), (5, 'Onion'), (6, 'Basil')
ON CONFLICT (id) DO NOTHING;

INSERT INTO dish_ingredients (dish_id, ingredient_id, removable)
VALUES
    (1, 1, TRUE), (1, 2, TRUE), (2, 3, TRUE), (2, 4, FALSE), (3, 5, TRUE), (4, 6, TRUE)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('shops', 'id'), COALESCE((SELECT MAX(id) FROM shops), 1));
SELECT setval(pg_get_serial_sequence('dishes', 'id'), COALESCE((SELECT MAX(id) FROM dishes), 1));
SELECT setval(pg_get_serial_sequence('ingredients', 'id'), COALESCE((SELECT MAX(id) FROM ingredients), 1));
