# Software Defined Vehicle End-to-End Demo
This end-to-end scenario demonstrates how to customize user’s driving experience using AI. When a user approaches a car, in-car computer vision detects the user’s face and retrieves the user’s profile. Then, based on the profile, the system adjusts both car interior ambient light and driver seat position to provide a tailored experience to the user.

When the car is staged in a dealership showroom, the system uses a in-store computer vision to detect user’s face and sends a user profile to the car for sales demonstration. The dealer can also customize on-car application packages and push selected applications to the car when the car is sold.

![E2E scenario](./images/e2e-scenario.png)

# Key Messages
* Leroy provides an edge-native programming model for in-car app developers to easily discover and incorporate various capabilities such as AI, pub-sub and state management.
* Leroy supports dynamically switching among different capability vendors based on policies and runtime telemetries such as network connectivity and performance. 
* Car middleware provides a shared platform for in-car applications to exchange messages, share state, and securely invoke each other.
* Symphony control plane allows cars to be automatically discovered and onboarded as deployment targets, to which in-car applications can be pushed from the control plane.
* Symphony supports resource projection to Azure to provide a central management experience across geographically distributed dealerships. 

# Demo Components

* **Face detection app** offers in-car face detection capability to detect user face.
* **Driving comfort app** reads user preferences and adjusts ambient lighting, as well as seat position for the driver.
* **Face detection service** provides face detection service in dealership showroom.
* **User profile service** provides templated user profiles for unregistered users. It also allows manual adjustments of settings like ambient light, temperature and seat position.

# Demo Preparation

The full-scale demonstration requires both software and hardware components. You can also run a simulated demonstration with only software components.

* [AI Model preparation](./docs/ai-preparation.md)
* [Hardware preparation](./docs/hardware-preparation.md)
* [Software preparation](./docs/software-preparation.md)


# Demo Scripts

* [Pre-demo steps](./docs/pre-demo-steps.md)
* [Demo steps](./docs/demo-steps.md)
* [Post-demo steps](./docs/post-demo-steps.md)