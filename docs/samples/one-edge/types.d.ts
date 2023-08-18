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
    spec: any;
}