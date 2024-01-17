# Campaign scenarios

A [campaign](./unified-object-model/campaign.md) is a versatile type that can be used to implement many useful workflows. This section contains a few typical use cases of campaigns.

For more information about how Symphony approaches workflows, see [Workflows](./workflows.md).

## Canary deployment

Symphony offers several ways to model a canary deployment, depending on your specific requirements. The following discussion assumes you have an application with a frontend and a backend, and you want to roll out a canary deployment for the backend. 

### Model the application

You have several options to model your application:

* A single solution

  The easiest way to model your application is to use a single `solution` object that has three components: `frontend`, `backend(v1)`, and `ingress`. When you are ready to do a canary deployment of `backend(v2)`, you add the `backend(v2)` as a new component to your application and reconfigure the `ingress`.

* Two solutions

  You can use two separate `solution` objects to represent the frontend and the backend. This option may be preferred if the frontend and the backend are managed by different teams. The frontend contains just the frontend component, and the backend contains the `backend(v1)` component and an `ingress` configuration. When you need to deploy a new backend version, you add a `backend(v2)` to your backend solution and adjust your `ingress` settings.

* Solutions and a target

  You can move the `ingress` component to a `target` object, especially when the ingress is considered IT infrastructure and managed by IT pros instead of developers. In such cases, you can model your application with either a single `solution` or two `solutions` as above and keep the `ingress` component in the `target` object.

### Shift the traffic

You can adjust the traffic pattern by adjusting your `ingress` settings, for instance to send 10% of traffic to the new version. Then, based on validation results, you can gradually shift the traffic to the new version, or roll back to the original version if validation fails.

### Automate canary deployment with a campaign

You can factor deployments, validations, and ingress updates into a campaign to automate the canary deployments. Conceptually, the campaign contains the following stages:

1. Add `backend(v2)` to your solution object. This can be done at a stage using a `patch` provider that patches your existing solution to add a new component.
2.	Patch your ingress configuration to shift a portion of traffic to the new version. The workflow jumps to stage 4 if all traffic has been shifted to the new version. Or it jumps to stage 5 if all traffic has been reverted to the original version.
3.	Using a `http` provider, run tests against the new version. Depending on the output, you can go back to stage 2 with a new percentage configuration to adjust the traffic pattern.
4.	If the canary deployment succeeded, remove backend(v1) from the solution and the workflow stops.
5.	If the canary deployment failed, you can choose to remove backend(v2) and the workflow stops.

## Gated deployment

Symphony campaigns support gated deployments. You can add a stage that invokes an HTTP endpoint (such as an [Azure Logic Apps](https://learn.microsoft.com/azure/logic-apps/logic-apps-overview)) to trigger a custom approval flow – such as sending an email to an approver and waiting for the approver to click on an “Approve” button in the email using Office 365 features. When the external approval flow succeeds, the campaign progresses to the next stage, which is to deploy the `instance`.

