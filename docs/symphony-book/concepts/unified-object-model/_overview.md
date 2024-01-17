# Unified object model

Symphony defines a common object model that describes the full stack of intelligent solutions, from AI models to solutions to devices and sensors. Because these objects are defined as standard Kubernetes [custom resources](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/), you can use popular Kubernetes tools like [kubectl](https://kubernetes.io/docs/reference/kubectl/kubectl/) to manipulate these objects.

* [AI Model](./ai-model.md) (`model.ai.symphony`)
* [AI Skill](./ai-skill.md) (`skill.ai.symphony`)
* [Device](./device.md) (`device.fabric.symphony`)
* [Target](./target.md) (`target.fabric.symphony`)
* [Solution](./solution.md) (`solution.solution.symphony`)
* [Instance](./instance.md) (`instance.solution.symphony`)

## Typical workflows

Depending on your focus, consider the typical workflows described below: the AI workflow, the device workflow, or the solution workflow. At the end of each workflow, you want to create `instance` objects which represent running deployments of your intelligent edge solution.

### AI workflow

1. Create your AI model using tools of your choice.
2. Once you have the AI model file, register your AI model with Symphony as a `model` object.
3. Define AI `skill` objects that define processing pipelines. A processing pipeline reads data from a data source, applies one or more AI models (and other transformations), and sends inference results to designated outputs.

### Device workflow

1. Register your computational devices with Symphony as `target` objects. You can also specify desired runtime components, such as Symphony agent, in your target definition.
2. Manually register your non-computational devices, such as sensors and actuators, as `device` objects. You can also leverage projects like [Akri](https://github.com/project-akri/akri) to auto-discover devices.

### Solution workflow

1. Define your intelligent solution as a `solution` object, which consists of a number of component elements. A component is usually a container, and it may refer to AI `skill` objects in its properties.
2. Define an `instance` object that maps a `solution` to one or multiple `target` objects. Once the instance object is created, Symphony ensures that the impacted targets are updated according to the desired solution state and target state.

## Sample workflow

Assume that you are creating an intelligent edge solution that uses a website to show the number of cars passing an intersection each hour. The following workflow describes how to create and deploy such a solution using Symphony API.

1. Create or select a car detection model. Symphony comes with a model zoo, which contains a car detection model you can use.
2. Register the model as a Symphony `model` object.
3. Define a Symphony `skill` object that defines a pipeline:
    * Take input from a camera
    * Send frames to the car detection model
    * Collect inference result and send detection event to an output (such as IoT Hub or an HTTP endpoint)
4. Define a Symphony `solution` object that has a Docker container component that takes the `skill` as input and drives the inference process. Since Symphony provides a couple of such containers out-of-box, you don't need to create these containers yourself.
5. Create your website container and add it as a component of your `solution`.
6. Define a `target` that represents a computer to which you want to deploy your solution.
7. Define `device` objects for cameras you want to use. These `device` objects are associated with your `target` object through labeling.
8. Create an `instance` object that deploys your `solution` to your `target`.
