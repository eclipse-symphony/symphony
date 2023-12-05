/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

export function stateToString(num: number): string {
    switch (num) {
        case 200:
            return "Success";
        case 500:
            return "Failed";
        case 9994:
            return "Running";
        case 9995: 
            return "Paused";
        case 9996:
            return "Done";  
        case 9997:
            return "Delayed";      
        case 9998:
            return "Untouched";
        default:
            return "Unknown";
    }
  }