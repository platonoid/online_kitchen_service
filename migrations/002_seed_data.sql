INSERT INTO restaurants (slug, name, city, cuisine, delivery_eta_min, commission_rate, is_active)
VALUES
    ('chai-corner', 'Chai Corner', 'Moscow', 'Indian', 25, 0.08, TRUE),
    ('pastalab', 'PastaLab', 'Moscow', 'Italian', 30, 0.10, TRUE),
    ('sushi-metro', 'Sushi Metro', 'Moscow', 'Japanese', 22, 0.09, TRUE)
ON CONFLICT (slug) DO NOTHING;

INSERT INTO restaurant_items (restaurant_id, name, description, category, price_cents, is_available)
VALUES
    (1, 'Masala Chai', 'Strong black tea with cardamom and milk.', 'drinks', 250, TRUE),
    (1, 'Chicken Biryani', 'Fragrant rice with chicken, spices and herbs.', 'main', 420, TRUE),
    (1, 'Paneer Wrap', 'Grilled paneer with salad and house sauce.', 'main', 390, TRUE),
    (2, 'Margherita Pizza', 'Tomato sauce, mozzarella and basil.', 'pizza', 540, TRUE),
    (2, 'Spaghetti Carbonara', 'Creamy pasta with bacon and parmesan.', 'pasta', 620, TRUE),
    (2, 'Caesar Salad', 'Romaine, chicken, parmesan and dressing.', 'salad', 430, TRUE),
    (3, 'Salmon Roll Set', 'Classic salmon sushi set with miso soup.', 'rolls', 760, TRUE),
    (3, 'Vegan Maki Box', 'Fresh vegetables and avocado in rice rolls.', 'maki', 490, TRUE),
    (3, 'Edamame Bowl', 'Warm edamame with sesame and chili.', 'snacks', 310, TRUE)
ON CONFLICT DO NOTHING;
