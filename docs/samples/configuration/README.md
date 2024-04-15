# Configuration Management Samples

This directory contains samples for managing various degrees of configuration complexity.  It includes examples and instructions for the following scenarios, all based on the Kubernetes provider.  Each scenario is a self-contained example, deployable to an existing Kubernetes cluster.

* [Conditional configuration activation](./conditional-activation/README.md): This section provides guidance on enabling or disabling specific configuration sections based on a predetermined flag. It explains how to conditionally include sections in the final configuration object, ensuring that only relevant configurations are applied.

* [Reusable configuration segments](./reusable-segments/README.md): This part focuses on the modularization of configuration elements. It demonstrates how to isolate commonly used configuration sections into separate objects for easy inclusion across multiple deployments, such as different lines, sites, or regions.

* [Array merging in configurations](./array-merging/README.md): This focuses on the management of configurations that contain extensive lists which may vary slightly by location but share a core set. The documentation outlines a method for separating the common elements into a single configuration object and merging them with unique sets specific to each location.

* [Using context to assemble configuration](./context-based/README.md): This example shows how to use contextual information from an instance to dynamically change configuration.
