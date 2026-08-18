# Script stage provider

Script stage provider runs a bash or PowerShell script as a campaign stage. The script can live in a local folder or be downloaded from a remote URL; remote URLs are validated against the server-side security policy before downloading. On each execution, the stage inputs are written to a JSON file whose path is passed to the script as an argument, and the script writes its results to a second JSON file that becomes the stage outputs.

## Config

| Field | Value |
|-------|-------|
| `name` | Provider name. |
| `script` | Script file name. |
| `scriptFolder` | (optional) Folder containing the script — a local path or a remote URL. |
| `stagingFolder` | (optional) Folder for the generated input/output files (and the downloaded script, when remote). |
| `scriptEngine` | (optional) `bash` (default) or `powershell`. |

## Inputs

| Field | Value |
|-------|-------|
| `<any field>` | All inputs are serialized to a JSON file that is passed to the script as its argument. |

## Outputs

| Field | Value |
|-------|-------|
| `<any field>` | The contents of the JSON output file written by the script, as key-value pairs. |

## Sample

Run an approval script downloaded from a remote folder, and branch on the `status` value it returns:

```yaml
approval:
  name: "approval"
  provider: "providers.stage.script"
  config:
    scriptFolder: "https://raw.githubusercontent.com/eclipse-symphony/symphony/main/docs/samples/approval/script"
    scriptEngine: "bash"
    script: "mock-reject.sh"
  stageSelector: ${{$if($equal($output(approval,status), 200),'deploy','end')}}
```
