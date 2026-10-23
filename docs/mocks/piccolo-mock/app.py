#
# Copyright (c) Microsoft Corporation.
# Licensed under the MIT license.
# SPDX-License-Identifier: MIT
#
from flask import Flask, request, jsonify

app = Flask(__name__)

# Mock storage for scenarios: keyed by workload_type, with each value being a list of workload names
# scenarios = {"binary": ["plc-controller"]}
scenarios = []

@app.route('/scenario/<workload_name>', methods=['GET'])
def get_workload(workload_name):
    """
    This endpoint simulates getting a workload from Piccolo.
    """
    # Search for workload_name within the specified workload_type
    if workload_name in scenarios:
        return jsonify({"message": "Workload found"}), 200
    else:
        return jsonify({"message": "Workload not found"}), 404

@app.route('/scenario', methods=['POST'])
def create_scenario():
    """
    This endpoint simulates creating a scenario.
    """
    # Extract workload_name from request body, or default to ""
    workload_name = request.data.decode('utf-8') or ''

    # Add component_name to the appropriate workload_type list if it doesn't already exist
    if workload_name not in scenarios:
        scenarios.append(workload_name)

    return jsonify({"message": f"Scenario '{workload_name}' created successfully"}), 201

@app.route('/scenario/<workload_name>', methods=['DELETE'])
def delete_scenario(workload_name):
    """
    This endpoint simulates deleting a scenario.
    """
    # Find and delete the component_name from all workload_type lists
    if workload_name in scenarios:
        scenarios.remove(workload_name)
        return jsonify({"message": f"Scenario '{workload_name}' deleted successfully"}), 200
    else:
        return jsonify({"error": f"Scenario '{workload_name}' not found"}), 404

if __name__ == '__main__':
    app.run(port=5000, debug=True)
