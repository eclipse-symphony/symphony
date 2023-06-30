import * as path from "node:path";
import { readFile } from "node:fs/promises";
import {
  IOTPipelineType,
  IOTPipelineGeneratorParams,
  IOTPipelineParams,
  IOTPipelineParamsMicroservice,
  IOTPipelineACRTenant,
  IOTPipelineHelmConfig,
  IOTPipelineHelmACR,
  IOTPipelineParamsHelmChart,
  IOTPipelineParamsBuildScript,
} from "@azure-iot/pipelines";
interface PipelineYaml {
  [key: string]: any;
}

const BYO_ACR_REGISTRY_DEV: string = "symphonycr.azurecr.io";

const BYO_ACR_SERVICE_CONNECTION_NAME_DEV: string = "symphony_corp_acr";

const DH_SERVICE_CONNECTION_NAME: string = "symphony_ado_conn";

export async function getIotPipelineParams(): Promise<IOTPipelineGeneratorParams> {
  const iotPipelineParams: IOTPipelineParams[] = [];

  iotPipelineParams.push(await getBuddyPipelineParams());
  iotPipelineParams.push(await getDevPipelineParams());
  iotPipelineParams.push(getPRPipelineParams());

  return {
    iotPipelineCommonParams: {
      buildHomeDirectory: ".setup/pipelines",
    },
    iotPipelineParams,
  };
}

async function getBuddyPipelineParams(): Promise<IOTPipelineParams> {
  return {
    type: IOTPipelineType.Buddy,
    buildTagIdentifier: "buddy",
    parallelJobCount: 1,
    showBuildServiceToggleOptions: false,
    microservices: getAllMicroservices(),
    acr: getDevSymphonyACR(),
    helmConfig: getDevHelmConfig(),
    customPrebuildScripts: await getCustomPrebuildScripts(),
    editHook: await generateCustomPipelineHook("buddy"),
  };
}

async function getDevPipelineParams(): Promise<IOTPipelineParams> {
  return {
    type: IOTPipelineType.Buddy,
    buildTagIdentifier: "develop",
    parallelJobCount: 1,
    trigger: {
      batch: true,
      branches: ["main"],
    },
    showBuildServiceToggleOptions: false,
    microservices: getAllMicroservices(),
    acr: getDevSymphonyACR(),
    helmConfig: getDevHelmConfig(),
    customPrebuildScripts: await getCustomPrebuildScripts(),
    editHook: await generateCustomPipelineHook("develop"),
  };
}

function getPRPipelineParams(): IOTPipelineParams {
  return {
    type: IOTPipelineType.PullRequest,
    buildTagIdentifier: "pr",
    parallelJobCount: 2,
    showBuildServiceToggleOptions: false,
    microservices: getAllMicroservices(),
    skipPrebuildCredsSetup: true,
    editHook: editPrPipeline,
  };
}

function getAllMicroservices(): IOTPipelineParamsMicroservice[] {
  const microservices: IOTPipelineParamsMicroservice[] = [];

  microservices.push(
    {
      repositoryName: "symphony-api",
      dockerFilePath: "/api/Dockerfile",
      dockerFileContextPath: "/",
    },
    {
      repositoryName: "symphony-k8s",
      dockerFilePath: "/k8s/Dockerfile",
      dockerFileContextPath: "/",
    }
  );

  return microservices;
}

function getDevSymphonyACR(): IOTPipelineHelmACR {
  return {
    tenant: IOTPipelineACRTenant.MSFT,
    registry: BYO_ACR_REGISTRY_DEV,
    serviceConnectionName: BYO_ACR_SERVICE_CONNECTION_NAME_DEV,
  };
}

function getDevHelmConfig(): IOTPipelineHelmConfig {
  return {
    helmAcr: getDevSymphonyACR(),
    helmCharts: getDevHelmCharts(),
  };
}

function getDevHelmCharts(): IOTPipelineParamsHelmChart[] {
  return [
    {
      helmChartPath: "/symphony-extension/helm/symphony",
    },
  ];
}

function editPrPipeline(pipelineYaml: PipelineYaml): PipelineYaml {
  return splitJobs(pipelineYaml);
}
// Note: This is a hack. We should be able to specify the number of jobs in the pipeline
// and have the steps distributed evenly across them. However, the pipeline generator
// currently only supports a single job. This function splits the steps across multiple
// jobs to allow for parallelization.
function splitJobs(pipelineYaml: PipelineYaml): PipelineYaml {
  const jobs = pipelineYaml.extends.parameters.stages[0].jobs;
  const jobCount = jobs.length;
  const steps = jobs[0].steps;
  // distribute steps evenly across jobs
  for (let i = 0; i < jobCount; i++) {
    jobs[i].steps = steps.slice(
      Math.floor((i * steps.length) / jobCount),
      Math.floor(((i + 1) * steps.length) / jobCount)
    );
  }
  return pipelineYaml;
}

async function generateCustomPipelineHook(name: string) {
  name = name.trim().replace(/\s+/g, "-").toLowerCase()
  const pipelineName = `${name}-$(Date:yyyyMMdd)$(Rev:.r)`;
  
  const inlineScript = await readFile(
    path.join(__dirname, "../scripts/lock-artifacts.sh"),
    "utf8"
  );

  return (pipelineYaml: PipelineYaml): PipelineYaml => {
    pipelineYaml.name = pipelineName;
    const finalJob = pipelineYaml.extends.parameters.stages.at(-1).jobs.at(-1);
    const task = {
      task: "AzureCLI@2",
      displayName: "Lock artifacts in ACR",
      condition: "ne(variables['Agent.OS'], 'Windows_NT')",
      inputs: {
        azureSubscription: DH_SERVICE_CONNECTION_NAME,
        scriptType: "bash",
        scriptLocation: "inlineScript",
        inlineScript,
        workingDirectory: "$(Build.SourcesDirectory)/dst/drop_prebuild_prebuild",
      },
    };
    finalJob.steps.push(task);
    return pipelineYaml;
  };
}

async function getCustomPrebuildScripts(): Promise<
  IOTPipelineParamsBuildScript[]
> {
  const script = await readFile(
    path.join(__dirname, "../scripts/prepare-chart.sh"),
    "utf8"
  );
  return [
    {
      type: "GoTool@0",
      displayName: "Install Go 1.20",
      inputs: {
        version: "1.20",
      },
    },
    {
      type: "Go@0",
      displayName: "Install install mage",
      inputs: {
        command: "install",
        arguments: "github.com/magefile/mage@latest",
      },
    },
    {
      type: "Bash@3",
      displayName: "Prepare Helm Chart",
      inputs: {
        targetType: "inline",
        script,
      },
    },
  ];
}
