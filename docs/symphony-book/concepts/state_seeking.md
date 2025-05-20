# State seeking

Many software management problems can be viewed as a state seeking problem: a system reports its current state, a user specifies a new desired state, and a state seeking system brings the current state towards the desired state.

With this high-level abstraction, Symphony unifies the workflow of software deployment, software update, configuration management, policy management, device update, firmware update, OS update and many other tasks where a system needs to be updated according to a given new state.

When Symphony defines state seeking, it identifies five concerns:

* **Understanding** – how the current state is understood
* **Authoring** – how the desired state is designed and authored
* **Staging** – how the desired state is staged for transmission to targets
* **Distribution** – how the desired state is transmitted from source to targets
* **Application** – how the desired state is applied to a target

Symphony doesn’t bet on any specific technology or artifact formats. Instead, it defines abstract API for each of the concerns and allows different systems to be plugged into the same state seeking flow.

For example, consider state distribution within Symphony. Symphony maintains a neutral stance regarding the method of distribution, whether it involves direct downloads, file copying, messaging protocols, or even manual distribution via USB drives. In its role as an orchestrator, Symphony's primary function is to initiate the transmission process by signaling the configured systems and subsequently awaiting their reports on the transmission status. All other intricacies and details of the distribution process are delegated to the respective systems involved. This indifference to implementation details allows Symphony to coordinate many artifact staging, shipping and application services and put them into a consistent experience, regardless of the protocols and formats they are using.
