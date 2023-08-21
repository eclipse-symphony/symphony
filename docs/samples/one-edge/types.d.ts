export interface Site {
    id: string;
    name: string;
    address: string,
    phone: string;
    city: string,
    sate: string,
    zip: string,
    country: string,
    version: string,
    self: boolean,
    lastReported: Date,
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
}
export interface CampaignSpec {
    id: string;
    firstStage: string;
    stages: Record<string, StageSpec>;
}
export interface Catalog {
    id: string;
    origin: string;
    type: string;
    name: string;
    properties: Record<string, any>;
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
