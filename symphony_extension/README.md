# Symphony Extension Configurations
Everything under the symphony_extension directory is documentation for the symphony arc extension.

## extension-registration-payloads
This directory contains the configurations used to register the Extension Types in the geneva portal.

## helm-charts
The modified symphony helm chart currently deployed to (dogfood) for the symphony extension. This chart was modified such that it can work in the dogfood environment. Further investigation needs to be performed to update original helm chart with the updates that were included in this one. When those updates get made then this helm-chart can be removed.

## resource-sync-rules
This md file contains the configuration and output of resource sync rules. The subscription, resource group, and custom location can be substituted with the users configurations. The endpoint in the configuration json is for dogfood.