'use client';

import {Card, CardHeader, CardBody, CardFooter, Divider, Link, Image} from '@nextui-org/react';
import { Catalog } from '../app/types';
import PropertyTable from './PropertyTable';
import SolutionCard from './SolutionCard';
import { FcSettings, FcTemplate } from 'react-icons/fc';

interface CatalogCardProps {
    catalog: Catalog;
}
function CatalogCard(props: CatalogCardProps) {
    const { catalog } = props;
    return (
        <Card>
            <CardHeader className="flex gap-3">
                {catalog.type === 'config' && (
                    <FcSettings className="text-[#F6B519] text-3xl"/>
                )}
                {catalog.type === 'solution' && (
                    <FcTemplate className="text-[#F6B519] text-3xl"/>
                )}
                <div className="flex flex-col">
                <p className="text-md">{catalog.name}</p>
                <p className="text-small text-default-500">{catalog.type}</p>
                </div>
            </CardHeader>
            <Divider/>
            <CardBody>
                {catalog.type === 'config' && (
                    <PropertyTable properties={catalog.properties} />
                )}
                {catalog.type === 'solution' && (
                    <SolutionCard solution={{
                        spec: catalog.properties['spec']
                    }} />
                )}
            </CardBody>
            {/* <Divider/>
            <CardFooter>
                    <span>MISS</span>
            </CardFooter> */}
        </Card>
    );
}
export default CatalogCard;