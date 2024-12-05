import dash
from dash import dcc, html
from dash.dependencies import Input, Output, State
import pymysql
import os
from datetime import datetime
import plotly.express as px

refresh_interval = 15000  # 15 seconds

def setup_db_table():
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
            time VARCHAR(255) NOT NULL,
            customerName VARCHAR(255) NOT NULL,
            items TEXT NOT NULL,
            status VARCHAR(50) NOT NULL
        )"""
    )
    cursor.execute(
        """CREATE TABLE IF NOT EXISTS item_count (
            time VARCHAR(255) PRIMARY KEY,
            Cookie INT NOT NULL,
            FruitSmoothie INT NOT NULL,
            Hotdog INT NOT NULL,
            IceCream INT NOT NULL,
            Nachos INT NOT NULL,
            PizzaSlice INT NOT NULL,
            Soda INT NOT NULL
        )"""
    )
    cursor.execute(
        """CREATE TABLE IF NOT EXISTS order_count (
            time VARCHAR(255) PRIMARY KEY,
            count INT NOT NULL
        )"""
    )
    conn.commit()
    conn.close()
    insert_initial_order_count()

def insert_initial_order_count():
    conn = pymysql.connect(
        host=os.getenv("MYSQL_HOST", "localhost"),
        user="root",
        database="ordersDB",
    )
    cursor = conn.cursor()
    try:
        conn.begin()
        # Check if there is already a record with order count is 0
        current_time = datetime.now().strftime('%Y-%m-%d %H:%M:%S')
        cursor.execute("SELECT * FROM order_count ORDER BY time DESC LIMIT 1")
        rows = cursor.fetchall()
        if len(rows) == 0:
            cursor.execute("INSERT INTO order_count (time, count) VALUES (%s, %s)", (current_time, 0))
        cursor.execute("SELECT * FROM item_count ORDER BY time DESC LIMIT 1")
        rows = cursor.fetchall()
        if len(rows) == 0:
             cursor.execute("INSERT INTO item_count (time, Cookie, FruitSmoothie, Hotdog, IceCream, Nachos, PizzaSlice, Soda) VALUES (%s, %s, %s, %s, %s, %s, %s, %s)", (current_time, 0, 0, 0, 0, 0, 0, 0))
        # Commit the transaction
        conn.commit()
    except Exception as e:
        # Rollback the transaction in case of error
        conn.rollback()
        print(f"Error: {e}")
    finally:
        conn.close()

# Function to fetch data from the database
def fetch_orders_count_data():
    setup_db_table()
    conn = pymysql.connect(
        host=os.getenv("MYSQL_HOST", "localhost"),
        user="root",
        database="ordersDB",
    )
    cursor = conn.cursor()
    cursor.execute("SELECT time, count FROM order_count ORDER BY time DESC LIMIT 10")
    rows = cursor.fetchall()
    conn.close()
    if len(rows) == 0:
        return None, None
    x_values = [row[0] for row in rows]
    y_values = [row[1] for row in rows]
    return x_values, y_values

def fetch_orders_by_type_data():
    setup_db_table()
    conn = pymysql.connect(
        host=os.getenv("MYSQL_HOST", "localhost"),
        user="root",
        database="ordersDB",
    )
    cursor = conn.cursor()
    cursor.execute("SELECT Cookie, FruitSmoothie, Hotdog, IceCream, Nachos, PizzaSlice, Soda FROM item_count ORDER BY time DESC LIMIT 1")
    row = cursor.fetchone()
    conn.close()
    if row:
        cookie_count = row[0]
        fruit_smoothie_count = row[1]
        hotdog_count = row[2]
        ice_cream_count = row[3]
        nachos_count = row[4]
        pizza_slice_count = row[5]
        soda_count = row[6]
        return cookie_count, fruit_smoothie_count, hotdog_count, ice_cream_count, nachos_count, pizza_slice_count, soda_count
    else:
        return 0, 0, 0, 0, 0, 0, 0

def create_dash_by_orders(flask_app, name):
    dash_app = dash.Dash(__name__, server=flask_app, url_base_pathname="/" + name + "/")
    
    # Dash layout
    dash_app.layout = html.Div([
        dcc.Graph(
            id=name,
            config={"displayModeBar": False},  # Hide toolbar
            style={"width": "100%", "height": "100%"}  # adaptive size
        ),
        dcc.Interval(
            id='interval-component-orders',
            interval=refresh_interval,
            n_intervals=0
        )
    ], style={"width": "100%", "height": "100%", "display": "flex", "flex-direction": "column"})

    # Dash callback
    @dash_app.callback(
        Output(name, "figure"),
        [Input('interval-component-orders', 'n_intervals')]
    )
    def update_graph(n_intervals):
        # Fetch data from the database
        x_values, y_values = fetch_orders_count_data()
        fig = px.line(x=x_values, y=y_values, labels={'x': 'time', 'y': 'Orders Count'})
        fig.update_layout(
            showlegend=False,  # Hide legend
            title=None,  # Hide title
            xaxis_title=None,  # Remove x-axis title
            yaxis_title=None,  # Remove y-axis title
            autosize=True,
            height=160,
            margin=dict(l=20, r=20, t=40, b=40),
        )
        return fig

    return dash_app

def create_dash_by_products(flask_app, name):
    dash_app = dash.Dash(__name__, server=flask_app, url_base_pathname="/" + name + "/")
    
    # Dash layout
    dash_app.layout = html.Div([
        dcc.Graph(
            id=name,
            config={"displayModeBar": False},  # Hide toolbar
            style={"width": "100%", "height": "100%", "overflow": "hidden"}  # adaptive size and hide overflow
        ),
        dcc.Interval(
            id='interval-component-products',
            interval=refresh_interval,
            n_intervals=0
        )
    ], style={"width": "100%", "height": "100%", "display": "flex", "flex-direction": "column"})

    # Dash callback
    @dash_app.callback(
        Output(name, "figure"),
        [Input('interval-component-products', 'n_intervals')]
    )
    def update_graph(n_intervals):
        # Fetch data from the database
        cookie_count, fruit_smoothie_count, hotdog_count, ice_cream_count, nachos_count, pizza_slice_count, soda_count = fetch_orders_by_type_data()
        
        # Check if data is fetched correctly
        if cookie_count is None or fruit_smoothie_count is None or hotdog_count is None or ice_cream_count is None or nachos_count is None or pizza_slice_count is None or soda_count is None:
            return dash.no_update
        
        # Create bar chart
        fig = px.bar(
            x=[cookie_count, fruit_smoothie_count, hotdog_count, ice_cream_count, nachos_count, pizza_slice_count, soda_count],
            y=['Cookie', 'Fruit Smoothie', 'Hotdog', 'Ice Cream', 'Nachos', 'Pizza Slice', 'Soda'],
            color=['Cookie', 'Fruit Smoothie', 'Hotdog', 'Ice Cream', 'Nachos', 'Pizza Slice', 'Soda'],
            color_discrete_map={'Cookie': 'red', 'Fruit Smoothie': 'orange', 'Hotdog': 'yellow', 'Ice Cream': 'green', 'Nachos': 'blue', 'Pizza Slice': 'purple', 'Soda': 'black'}
        )
        fig.update_layout(
            showlegend=False,  # Hide legend
            title=None,  # Hide title
            margin=dict(l=20, r=20, t=20, b=20),
            autosize=True,
            height=160,
            xaxis_title=None,  # Remove x-axis title
            yaxis_title=None,  # Remove y-axis title
        )
        return fig

    return dash_app