import * as path from "path";
import { IOTPipelineGenerator } from "@azure-iot/pipelines";
import { getIotPipelineParams } from "./pipelinesHelpers";

async function main() {
  // define path root level .pipelines directory, pipeline yaml will be written here
  const pipelineDefinitionsPath = path.join(__dirname, "../../../.pipelines/");

  // initialize IOTPipelineGenerator with IOTPipelineGeneratorParams
  const generator = new IOTPipelineGenerator(await getIotPipelineParams());

  // generate/write pipeline yaml files
  generator.writeAllPipelines(pipelineDefinitionsPath);
}

main();
