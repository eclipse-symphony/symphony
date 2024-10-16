from flask import Flask, request, jsonify

app = Flask(__name__)

# Mock storage for scenarios: keyed by workload_type, with each value being a list of workload names
scenarios = {"binary": ["plc-controller"]}

@app.route('/<workload_type>/<workload_name>', methods=['GET'])
def get_workload(workload_type, workload_name):
    """
    This endpoint simulates getting a workload from Piccolo.
    """
    # Search for workload_name within the specified workload_type
    if workload_type in scenarios and workload_name in scenarios[workload_type]:
        return jsonify({"message": "Workload found"}), 200
    else:
        return jsonify({"message": "Workload not found"}), 404

@app.route('/create-scenario/<component_name>', methods=['POST'])
def create_scenario(component_name):
    """
    This endpoint simulates creating a scenario.
    """
    req_body = request.get_json() or {}

    # Extract workload_type from request body, or default to "binary"
    workload_type = req_body.get("properties", {}).get("workload_type", "binary")

    # Add workload to scenarios dictionary
    if workload_type not in scenarios:
        scenarios[workload_type] = []

    # Add component_name to the appropriate workload_type list if it doesn't already exist
    if component_name not in scenarios[workload_type]:
        scenarios[workload_type].append(component_name)

    return jsonify({"message": f"Scenario '{component_name}' of type '{workload_type}' created successfully"}), 201

@app.route('/delete-scenario/<component_name>', methods=['DELETE'])
def delete_scenario(component_name):
    """
    This endpoint simulates deleting a scenario.
    """
    # Find and delete the component_name from all workload_type lists
    for workload_type, workload_list in scenarios.items():
        if component_name in workload_list:
            workload_list.remove(component_name)
            return jsonify({"message": f"Scenario '{component_name}' deleted successfully"}), 200

    return jsonify({"error": f"Scenario '{component_name}' not found"}), 404

if __name__ == '__main__':
    app.run(port=5000, debug=True)
