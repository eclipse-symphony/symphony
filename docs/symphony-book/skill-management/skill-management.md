# AI skill management

Symphony’s AI `skill` object is a generic graph definition with edges and nodes. The goal of this graph is not to replace existing AI graphs. Instead, it’s specifically aimed at creating a “super graph” that spans across multiple graphs from different platforms. Within this object, a graph from a specific AI framework can be treated as a single node, such as a DeepStream node. Symphony’s skill object augments that graph with:

* Consistent parameter handling
* Standardized camera discovery and auto-configuration
* Normalized import/export nodes and connectivity to business logic

The following diagram illustrates a sample integration with a DeepStream graph. The AI skill definition contains a camera input node, a DeepStream graph node, and an export node. This AI skill is then packaged into an AI skill package through a builder. The builder translates Symphony camera definition, along with any user specified parameters, into DeepStream native format. A Symphony adapter on the other end translates DeepStream output to Symphony format and feeds the output to a Symphony output node.

![ai-skill](../images/ai-skill.png)

This design allows AI vendors to bring in their AI graphs as an AI skill node of a Symphony AI skill graph. The previous example shows the entire graph being encapsulated as a single node. However, a vendor can also choose to supply individual nodes at finer granularity. The Symphony builder inserts adapters as needed to allow nodes from different vendors to communicate with each other when possible (configuration A in the following diagram). And the adapters are skipped among nodes from the same vendor to minimize overhead (configuration B). Finally, specialized vendor-to-vendor adapters are also allowed (configuration C).

![ai-skill-nodes](../images/ai-skill-nodes.png)

## Related topics

* [AI skill schema](../concepts/unified-object-model/ai-skill.md)