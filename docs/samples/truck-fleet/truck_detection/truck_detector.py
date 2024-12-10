import cv2
import torch
import time
import threading
import requests
import json

symphony_base_url = "http://localhost:8080/v1alpha2"

# Load the YOLO model
model = torch.hub.load('ultralytics/yolov5', 'yolov5s', pretrained=True)  # Use a pre-trained YOLOv5 model

# Configurable object type
object_to_detect = "truck"  # Change this to "truck", "bottle", etc., as needed

# Set the webcam index (0 is usually the default camera)
camera_index = 0

# Open the webcam
cap = cv2.VideoCapture(camera_index)

if not cap.isOpened():
    print("Error: Unable to access the camera")
    exit()

print("Press 'q' to quit the application.")

# Function to send the POST request in the background
def send_truck_detected_request():
    print("Truck detected for the specified duration, sending POST request.")

    try:
        # Authenticate and get the token
        auth_response = requests.post(
            f"{symphony_base_url}/users/auth",
            headers={"Content-Type": "application/json"},
            data=json.dumps({"username": "admin", "password": ""})
        )
        auth_response.raise_for_status()  # Raise an error for bad status codes
        token = auth_response.json().get("accessToken")

        print(f'Authenticated successfully. Token: {token}')

        # Read the campaign activation data from the JSON file
        with open('campaign_activation.json', 'r') as file:
            campaign_data = json.load(file)

        # Activate the campaign
        activation_response = requests.post(
            f"{symphony_base_url}/activations/registry/box-truck-docking-activation",
            headers={
                "Authorization": f"Bearer {token}",
                "Content-Type": "application/json"
            },
            data=json.dumps(campaign_data)
        )
        activation_response.raise_for_status()  # Raise an error for bad status codes

        print("Campaign activated successfully")
    except Exception as e:
        print(f"Error during request: {e}")

# Variables for truck detection logic
frame_count = 0
frames_to_confirm = 30  # Number of frames required to confirm detection
has_send = False 
# Main loop for real-time detection
while True:
    ret, frame = cap.read()
    if not ret:
        print("Error: Unable to read frame from camera")
        break

    # Perform inference
    results = model(frame)

    # Filter detections based on the object_to_detect
    filtered_results = results.pandas().xyxy[0]  # Get detection results as a Pandas DataFrame
    filtered_results = filtered_results[filtered_results['name'] == object_to_detect]

    # Check if a truck is detected
    if not filtered_results.empty:
        frame_count += 1

    # If truck detected in several consecutive frames, send request
    if frame_count >= frames_to_confirm and not has_send:
        threading.Thread(target=send_truck_detected_request).start()
        has_send = True

    # Draw bounding boxes for filtered detections
    detection_frame = frame.copy()
    for _, row in filtered_results.iterrows():
        x_min, y_min, x_max, y_max = int(row['xmin']), int(row['ymin']), int(row['xmax']), int(row['ymax'])
        confidence = row['confidence']
        label = f"{row['name']} {confidence:.2f}"

        # Draw rectangle and label
        cv2.rectangle(detection_frame, (x_min, y_min), (x_max, y_max), (0, 255, 0), 2)
        cv2.putText(detection_frame, label, (x_min, y_min - 10), cv2.FONT_HERSHEY_SIMPLEX, 0.5, (0, 255, 0), 2)

    # Display the original camera feed and the detection frame side by side
    cv2.imshow('Camera Preview and Detection', detection_frame)

    # Press 'q' to exit the loop
    if cv2.waitKey(1) & 0xFF == ord('q'):
        break

# Release resources
cap.release()
cv2.destroyAllWindows()
