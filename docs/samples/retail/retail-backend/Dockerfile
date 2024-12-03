# Use the official Python 3.9-slim-buster image as the base image
FROM python:3.9-slim-buster

# Set the working directory to /app
WORKDIR /app

# Copy the requirements.txt file to the container
COPY requirements.txt .

RUN pip install --upgrade pip

# Install the required packages
RUN pip install -r requirements.txt

# Copy the Flask application to the container
COPY . .

# Expose port 5000 for the Flask application
EXPOSE 5000

# Start the Flask application
CMD ["flask", "run", "--host", "0.0.0.0"]