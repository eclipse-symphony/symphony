'use client';

import {Card, CardHeader, CardBody, CardFooter, Divider, Link, Image} from '@nextui-org/react';
import { CatalogState } from '../../app/types';
import PropertyTable from '../PropertyTable';
import SolutionCard from '../SolutionCard';
import { FcSettings, FcTemplate } from 'react-icons/fc';

interface CatalogCardProps {
    catalog: CatalogState;
    refCatalog: CatalogState;
}
function CatalogCard(props: CatalogCardProps) {
    const { catalog, refCatalog } = props;
    return (
        <Card>
            <CardHeader className="flex gap-3">
                {catalog.spec.type === 'config' && (
                    <FcSettings className="text-[#F6B519] text-3xl"/>
                )}
                {catalog.spec.type === 'solution' && (
                    <FcTemplate className="text-[#F6B519] text-3xl"/>
                )}
                <div className="flex flex-col">
                <p className="text-md">{catalog.spec.name}</p>
                <p className="text-small text-default-500">{catalog.spec.type}</p>
                </div>
            </CardHeader>
            <Divider/>
            <CardBody>
                {catalog.spec.type === 'config' && (
                    <PropertyTable properties={catalog.spec.properties} refProperties={refCatalog?.spec.properties} />
                )}
                {catalog.spec.type === 'solution' && (
                    <SolutionCard solution={{
                        spec: catalog.spec.properties['spec']
                    }} />
                )}
            </CardBody>
            <Divider/>
            {catalog.spec.metadata?.['override'] && (
            
            <CardFooter>
                <span className='text-sm'>
                    <div>{`overrides: ${catalog.spec.metadata['override']}`}</div>                </span>
            </CardFooter>
            )}
        </Card>
    );
}
export default CatalogCard;