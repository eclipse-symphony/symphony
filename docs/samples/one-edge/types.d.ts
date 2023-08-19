export interface Site {
    id: string;
    name: string;
    description: string;
    phone: string;
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
