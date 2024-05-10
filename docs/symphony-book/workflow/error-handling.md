# Error Handling and Retries

A campaign activation stops running if any of the stages fails. However, in some cases you may want to run some error-handling logic when a stage fails. To mark a stage as an error-handler, you can annotate the stage with a `handleErrors` attribute. When a stage fails, its stage selector is evaluated. If the selected stage is an error-handling stage, then the error-handler is executed. Otherwise the activation fails. Once the error-handling stage is executed, the activation is allowed to continue as normal. 

The error-handling stage can be useful in many scenarios, such as sending a notification to a user if a deployment fails or retrying an operation several times.

