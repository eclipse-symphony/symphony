from flask import Flask, render_template, request, jsonify
import pymysql
import os
import json

app = Flask(__name__, static_folder="templates")

# Create Order
@app.route('/orders', methods=['POST'])
def create_order():
    order_data = request.json
    customer_name = order_data['customerName']
    items = order_data['items']
    status = 'Order Received'
    
    conn = pymysql.connect(
        host=os.getenv("MYSQL_HOST", "localhost"),
        user="root",
    )
    cursor = conn.cursor()

    # Create database if it does not exist
    cursor.execute("CREATE DATABASE IF NOT EXISTS ordersDB")
    cursor.execute("USE ordersDB")
    
    cursor.execute(
        """CREATE TABLE IF NOT EXISTS orders (
            orderId INT AUTO_INCREMENT PRIMARY KEY,
            customerName VARCHAR(255) NOT NULL,
            items TEXT NOT NULL,
            status VARCHAR(50) NOT NULL
        )"""
    )

    cursor.execute("INSERT INTO orders (customerName, items, status) VALUES (%s, %s, %s)", (customer_name, json.dumps(items), status))
    order_id = cursor.lastrowid
    
    conn.commit()
    conn.close()
    
    return jsonify({'orderId': order_id, 'customerName': customer_name, 'items': items, 'status': status}), 201

# Get Order
@app.route('/orders/<int:order_id>', methods=['GET'])
def get_order(order_id):
    conn = pymysql.connect(
        host=os.getenv("MYSQL_HOST", "localhost"),
        user="root",
        database="ordersDB",
    )
    cursor = conn.cursor()
    # Create database if it does not exist
    cursor.execute("CREATE DATABASE IF NOT EXISTS ordersDB")
    cursor.execute("USE ordersDB")
    
    cursor.execute(
        """CREATE TABLE IF NOT EXISTS orders (
            orderId INT AUTO_INCREMENT PRIMARY KEY,
            customerName VARCHAR(255) NOT NULL,
            items TEXT NOT NULL,
            status VARCHAR(50) NOT NULL
        )"""
    )

    cursor.execute("SELECT * FROM orders WHERE orderId = %s", (order_id,))
    order = cursor.fetchone()
    conn.close()
    
    if order:
        order = {
            'orderId': order[0],
            'customerName': order[1],
            'items': json.loads(order[2]),
            'status': order[3]
        }
        return jsonify(order)
    else:
        return jsonify({'error': 'Order not found'}), 404

# Update Order Status
@app.route('/orders/<int:order_id>/status', methods=['POST'])
def update_order_status(order_id):
    status_data = request.json
    new_status = status_data['status']
    
    conn = pymysql.connect(
        host=os.getenv("MYSQL_HOST", "localhost"),
        user="root",
        database="ordersDB",
    )
    cursor = conn.cursor()
    # Create database if it does not exist
    cursor.execute("CREATE DATABASE IF NOT EXISTS ordersDB")
    cursor.execute("USE ordersDB")
    
    cursor.execute(
        """CREATE TABLE IF NOT EXISTS orders (
            orderId INT AUTO_INCREMENT PRIMARY KEY,
            customerName VARCHAR(255) NOT NULL,
            items TEXT NOT NULL,
            status VARCHAR(50) NOT NULL
        )"""
    )

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

if __name__ == '__main__':
    app.run(debug=True)