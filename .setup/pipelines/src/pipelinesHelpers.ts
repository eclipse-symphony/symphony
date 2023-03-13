import { IOTPipelineType, IOTPipelineGeneratorParams, IOTPipelineParams, IOTPipelineParamsMicroservice, IOTPipelineACRTenant, IOTPipelineACR, IOTPipelineHelmConfig, IOTPipelineHelmACR, IOTPipelineParamsHelmChart } from '@azure-iot/pipelines'

/*
================================================================================================================================

getIotPipelineParams - entry point, returns IOTPipelineGeneratorParams

================================================================================================================================
*/

export function getIotPipelineParams(): IOTPipelineGeneratorParams {

    // array of individual pipeline params, each element will match to a generated yml file
    const iotPipelineParams: IOTPipelineParams[] = [];

    // push pipeline params for individual pipelines
    iotPipelineParams.push(getBuddyPipelineParams());

    // return IOTPipelineGeneratorParams
    return {
        // common params for all pipelines under this service
        iotPipelineCommonParams: {
            // npm registry to pull for this service
            npmRegistry: '@azure-iot:registry=https://pkgs.dev.azure.com/msazure/_packaging/AzureIOTSaas/npm/registry/',
            // set to path pipelines setup directory
            buildHomeDirectory: '.setup/pipelines',
            customPrebuildScripts: [],
        },
        // pipeline params
        iotPipelineParams,
    }
}

/*
================================================================================================================================

Pipeline Definitions

================================================================================================================================
*/

function getBuddyPipelineParams(): IOTPipelineParams {
    return {
        type: IOTPipelineType.Buddy,
        buildTagIdentifier: 'buddy',
        parallelJobCount: 1,
        showBuildServiceToggleOptions: false,
        microservices: getAllMicroservices(),
        acr: getIOTPipelineDevCorpACR(),
    }
}

/*
================================================================================================================================

Microservices/HelmChart Definitions

================================================================================================================================
*/

function getAllMicroservices(): IOTPipelineParamsMicroservice[] {
    const microservices: IOTPipelineParamsMicroservice[] = [];

    microservices.push({
        repositoryName: 'symphony-api', // ACR repo name for this service, ~service name
        dockerFilePath: '/api/Dockerfile', // repo path to Dockerfile to build
        buildByDefault: false,
    },
    {
        repositoryName: 'symphony-k8s', // ACR repo name for this service, ~service name
        dockerFilePath: '/k8s/Dockerfile', // repo path to Dockerfile to build
        buildByDefault: false,
    });

    return microservices;
}

/*
================================================================================================================================

ACR Definitions

================================================================================================================================
*/

// ACR info for where images should be pushed
const BYO_ACR_REGISTRY_DEV_CORP: string = 'symphonycr.azurecr.io';
// service connection name if needed, not necessary for pushes to isolated ACR
const BYO_ACR_SERVICE_CONNECTION_NAME_DEV_CORP: string = 'symphony_corp_acr';

// defining Docker build ACR to push to
function getIOTPipelineDevCorpACR(): IOTPipelineACR {
    return {
        tenant: IOTPipelineACRTenant.MSFT,
        registry: BYO_ACR_REGISTRY_DEV_CORP,
        serviceConnectionName: BYO_ACR_SERVICE_CONNECTION_NAME_DEV_CORP,
    };
}
