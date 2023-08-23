export function stateToString(num: number): string {
    switch (num) {
        case 9996:
            return "Done";
        case 9994:
            return "Running";
        case 9998:
            return "Untouched";
        default:
            return "Unknown";
    }
  }