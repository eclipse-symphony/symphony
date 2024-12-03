from flask import Flask, render_template, request, jsonify
import pymysql
import os
import json
from dash_app import create_dash_by_orders, create_dash_by_products, setup_db_table

app = Flask(__name__, static_folder="templates")
# Initialize Dash apps
dash_app_orders = create_dash_by_orders(app, "orders")
dash_app_products = create_dash_by_products(app, "products")

def parse_items(items):
    item_counts = {
        "Cookie": 0,
        "FruitSmoothie": 0,
        "Hotdog": 0,
        "IceCream": 0,
        "Nachos": 0,
        "PizzaSlice": 0,
        "Soda": 0
    }
    for item in items:
        name = item.get("name")
        quantity = item.get("quantity", 0)
        if name in item_counts:
            item_counts[name] += quantity
    return item_counts

@app.route('/')
def index():
    data_show = {
        "notifications": get_notifications(),
    }
    return render_template('index.html', data_show=data_show)

# Create Order
@app.route('/orders', methods=['POST'])
def create_order():
    order_data = request.json
    customer_name = order_data['customerName']
    items = order_data['items']
    order_time = order_data['time']
    status = 'Order Received'
    
    setup_db_table()

    conn = pymysql.connect(
        host=os.getenv("MYSQL_HOST", "localhost"),
        user="root",
        database="ordersDB",
    )
    cursor = conn.cursor()
    cursor.execute("INSERT INTO orders (customerName, time, items, status) VALUES (%s, %s, %s, %s)", (customer_name, order_time, json.dumps(items), status))
    order_id = cursor.lastrowid

    cursor.execute("SELECT time, count FROM order_count ORDER BY time DESC LIMIT 1")
    rows = cursor.fetchone()
    order_count = rows[1] + 1
    cursor.execute("INSERT INTO order_count (time, count) VALUES (%s, %s)", (order_time, order_count))

    item_counts = parse_items(items)
    cursor.execute("SELECT Cookie, FruitSmoothie, Hotdog, IceCream, Nachos, PizzaSlice, Soda FROM item_count ORDER BY time DESC LIMIT 1")
    rows = cursor.fetchone()
    cookie_count = rows[0] + item_counts["Cookie"]
    fruit_smoothie_count = rows[1] + item_counts["FruitSmoothie"]
    hotdog_count = rows[2] + item_counts["Hotdog"]
    ice_cream_count = rows[3] + item_counts["IceCream"]
    nachos_count = rows[4] + item_counts["Nachos"]
    pizza_slice_count = rows[5] + item_counts["PizzaSlice"]
    soda_count = rows[6] + item_counts["Soda"]
    cursor.execute("INSERT INTO item_count (time, Cookie, FruitSmoothie, Hotdog, IceCream, Nachos, PizzaSlice, Soda) VALUES (%s, %s, %s, %s, %s, %s, %s, %s)", (order_time, cookie_count, fruit_smoothie_count, hotdog_count, ice_cream_count, nachos_count, pizza_slice_count, soda_count))
    
    conn.commit()
    conn.close()
    
    return jsonify({'orderId': order_id, 'time': order_time, 'customerName': customer_name, 'items': items, 'status': status}), 201

# Get Order
@app.route('/orders/<int:order_id>', methods=['GET'])
def get_order(order_id):
    setup_db_table()
    conn = pymysql.connect(
        host=os.getenv("MYSQL_HOST", "localhost"),
        user="root",
        database="ordersDB",
    )
    cursor = conn.cursor()
    # Create database if it does not exist
    cursor.execute("CREATE DATABASE IF NOT EXISTS ordersDB")
    cursor.execute("USE ordersDB")

    cursor.execute("SELECT orderId, time, customerName, items, status FROM orders WHERE orderId = %s", (order_id,))
    order = cursor.fetchone()
    conn.close()
    
    if order:
        order = {
            'orderId': order[0],
            'time': order[1],
            'customerName': order[2],
            'items': json.loads(order[3]),
            'status': order[4]
        }
        return jsonify(order)
    else:
        return jsonify({'error': 'Order not found'}), 404

# Update Order Status
@app.route('/orders/<int:order_id>/status', methods=['POST'])
def update_order_status(order_id):
    status_data = request.json
    new_status = status_data['status']
    
    setup_db_table()

    conn = pymysql.connect(
        host=os.getenv("MYSQL_HOST", "localhost"),
        user="root",
        database="ordersDB",
    )
    cursor = conn.cursor()

    cursor.execute("SELECT * FROM orders WHERE orderId = %s", (order_id,))
    order = cursor.fetchone()
    if not order:
        conn.close()
        return jsonify({'error': 'Order not found'}), 404
    
    cursor.execute("UPDATE orders SET status = %s WHERE orderId = %s", (new_status, order_id))
    conn.commit()
    conn.close()
    
    return jsonify({'orderId': order_id, 'status': new_status})

@app.errorhandler(404)
def page_not_found(e):
    return render_template('404.html'), 404

def get_notifications():
    return [
        {"message": "Ordering Kiosk 1 is offline.", "type": "alarm", "color": "red"},
        {"message": "Order #1110 received.", "type": "warning", "color": "orange"},
        {"message": "Order #1109 received.", "type": "warning", "color": "orange"},
        {"message": "Order #1108 is ready for pickup.", "type": "info", "color": "green"},
    ]

if __name__ == '__main__':
    app.run(debug=True)