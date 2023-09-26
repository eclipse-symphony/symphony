# Workflows

Symphony employs a state-seeking approach to maintain system consistency in alignment with user-defined desired states. Nevertheless, there are situations where orchestrating system control necessitates capabilities beyond mere state-seeking. Take, for instance, the requirement for an approval process prior to initiating a deployment; in such cases, Symphony must possess the capability to trigger an approval workflow before proceeding with its state-seeking operations. Furthermore, for more complex deployment scenarios like canary deployments, multi-site rollouts, and staged deployments, Symphony must also offer robust workflow support.

Symphony’s workflow is described by a ```Campaign``` object. A Campaign can be activated by creating the corresponding ```Activation``` object. Each Activation object maintains its own state and context. Symphony uses an event-based approach to drive the Activation objects. And it allows steps in a Campaign to be executed remotely on a different Symphony control plane. A Campaign is comprised of one or multiple ```Stages```. I can be set to the default self-driving model, and it can also be put into a “stepped” model, which essentially transforms a Campaign into a collection of stateful nano services that can be invoked by an external workflow engine.

## Topics
* [Campaigns](./campaign.md)
* [Campaign Scenarios](./campaign-scenarios.md)
