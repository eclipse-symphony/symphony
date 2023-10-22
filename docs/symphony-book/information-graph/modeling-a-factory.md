# Modeling a Factory using the ISA95 Model

The following diagram illustrates a typical structure of an industrial automation system using the ISA95 model. Such a system encompasses multiple sub-systems, including ERP, MES, and SCADA, as well as various device categories like servers, PLCs, and sensors. Additionally, it involves multiple network connections, various user roles, policies, and numerous other components.

Using Symphony Catalogs, you can model various aspects of the system as graphs and link into existing data sources like ERP systems. And then, you can associate dynamic states of your software, hardware, configurations and policies with these graphs to gain end-to-end visibility of the entire system.

![isa-95](../images/isa-95.png)

Symphony offers several key platform-agnostic capabilities for modeling such a complex system:

* Modeling arbitrary information graph using Catalogs such as:
  * Asset trees
  * BOMs
  * Network topologies
  * Application templates
  * Configurations
  * Policies
* Modeling the entire stack of software on devices using Targets.
* Modeling applications using Solutions.
* Modeling application deployment topologies using Instances.
* Modeling AI pipelines using AI Skill.
* Modeling distributed workflows using Campaigns.

And because Symphony allows live states to be associated with the information graph, you can use these information graphs to gain live insights of your systems from different perspectives of your choice.
