# Backend Service Design

## Architecture Overview

The backend service will handle order processing and status updates, providing APIs for the frontend UWP app and the kitchen staff interface.

## Components

1. Order Management Service: This service will handle the creation, updating, and retrieval of orders. It will provide RESTful APIs for the UWP app to place orders and for the kitchen staff to update order statuses.

2. Database: A relational database (e.g., MySQL) will store order details, including customer information, order items, and status updates.

## Workflow

1. Order Placement: When a customer places an order using the UWP app, the order details are sent to the Order Management Service via a RESTful API.

2. Backend Processing: The Order Management Service stores the order details in the database.

3. Order Stages: The kitchen staff can view and update the order status through their interface to the Order Management Service, which updates the status in the database.

## RESTful API Endpoints

Create/Get/Update methods are implemented in this retail backend app but only Create method is used now. We can keep those remaining methods as follow up.

### 1. Create Order

Endpoint: POST `/orders`

Description: Creates a new order.

Request Body:

```javascript
{
  "customerName": "test",
  "items": [
    {"name": "coffee", "quantity": 1},
    {"name": "sandwich", "quantity": 1}
  ],
  "time": "2024-11-29 13:21:20"
}
```

Response:

```javascript
{
  "orderId": "1001",
  "status": "Order Received"
}
```

### 2. Get Order

Endpoint: GET `/orders/{orderId}`

Description: Retrieves the details of an order.

Response:

```javascript
{
  "orderId": "1001",
  "customerName": "test",
  "items": [
    {"name": "coffee", "quantity": 1},
    {"name": "sandwich", "quantity": 1}
  ],
  "time": "2024-11-29 13:21:20",
  "status": "order received"
}
```

### 3. Update Order Status

Endpoint: POST `/orders/{orderId}/status`

Description: Updates the status of an order.

Request Body:

```javascript
{
  "orderId": "1001",
  "status": "preparing"
}
```

Response:

```javascript
{
  "orderId": "1001",
  "status": "preparing"
}
```
