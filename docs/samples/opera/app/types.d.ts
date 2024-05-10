/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
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
    namespace: string;
    address: string;
    generation: string;
    metadata: Record<string, string>;
}

export interface CatalogSpec {
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

export interface SolutionState {
    id: string; 
    namespace: string;
    spec: SolutionSpec;
}

export interface SolutionSpec {
    displayName: string;
    components: ComponentSpec[];
    metadata: Record<string, string>;    
}

export interface ComponentSpec {
    name: string;
    type: string;
    properties: Record<string, any>;
    metadata: Record<string, string>;
    parameters: Record<string, string>;
    routes: RouteSpec[];
    constraints: string;
    depedencies: string[];
    skills: string[];
    sidecars: SidecarSpec[];
}

export interface RouteSpec {
    route: string;
    type: string;
    properties: Record<string, string>;
    filters: filterSpec[];
}

export interface filterSpec {
    direction: string;
    type: string;
    parameters: Record<string, string>;
}

export interface SidecarSpec {
    name: string;
    type: string;
    properties: Record<string, any>;
}

export interface GroupInfo {
    catalogs: Catalog[];
    title: string;
    type: string;
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