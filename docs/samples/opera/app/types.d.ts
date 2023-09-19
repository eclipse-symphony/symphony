/*
MIT License

Copyright (c) Microsoft Corporation.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE
*/

export interface Site {
    id: string;
    name: string;
    address: string,
    phone: string;
    city: string,
    state: string,
    zip: string,
    country: string,
    version: string,
    self: boolean,
    lastReported: Date,
    lat: number,
    lng: number,
}

export interface CampaignState {
    id: string;
    spec: CampaignSpec;
    status: CampainStatus;
}

export interface CampainStatus {

}

export interface StageSpec {
    name: string;    
    stageSelector: string;
    provider: string;
    inputs: Record<string, any>;
    config: Record<string, any>;
    contexts: string;
}

export interface CampaignSpec {
    id: string;
    firstStage: string;
    stages: Record<string, StageSpec>;
}

export interface ObjectRef {
	siteId: string;
    name: string;
    group: string;
    version: string;
    kind: string;
    scope: string;
    address: string;
    generation: string;
    metadata: Record<string, string>;
}

export interface CatalogSpec {
    siteId: string;
    name: string;
    type: string;
    properties: Record<string, any>;
    metadata: Record<string, string>;
    parentName: string;
    objectRef: ObjectRef;
    generation: string;
}

export interface CatalogStatus{
    properties: Record<string, string>;
}

export interface CatalogState {
    id: string;
    spec: CatalogSpec;
    status: CatalogStatus;
}

export interface GroupInfo {
    catalogs: Catalog[];
    title: string;
    type: string;
}

export interface ComponentSpec {
    name: string;
    type: string;
    properties: Record<string, any>;
}

export interface SolutionSpec {
    displayName: string;
    components: ComponentSpec[];
}

export interface Solution {
    id: string;
    spec: SolutionSpec;
}

export interface ActivationSpec {
    campaign: string;
    name: string;
    stage: string;
    inputs: Record<string, any>;
    generation: string;
}

export interface ActivationStatus {
    stage: string;
    nextStage: string;
    inputs: Record<string, any>;
    outputs: Record<string, any>;
    status: number;
    errorMessage: string;
    isActive: boolean;
    activationGeneration: string;
}

export interface ActivationState {
    id: string;
    metadata: Record<string, string>;
    spec: ActivationSpec;
    status: ActivationStatus;
}

export interface User {
    accessToken?: string;
    name?: string | null | undefined;
    email?: string | nulll | undefined;
    image?: string | null | undefined;
    username?: string;
    tokenType: string;
    roles?: string[] | undefined;
}

export interface Rule {
    type? : string;
    required?: boolean;
    pattern?: string;
    expression?: string;
}
export interface Schema {
    rules: Record<string, Rule>;
}