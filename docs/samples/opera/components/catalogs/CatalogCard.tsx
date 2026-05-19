'use client';

import {Card, CardHeader, CardBody, CardFooter, Divider, Link, Image} from '@nextui-org/react';
import { CatalogVersionState } from '../../app/types';
import PropertyTable from '../PropertyTable';
import SolutionVersionSpecCard from '../SolutionVersionSpecCard';
import { FcSettings, FcTemplate } from 'react-icons/fc';
import { FaGithub } from 'react-icons/fa';

interface CatalogVersionCardProps {
    catalogversion: CatalogVersionState;
    refCatalogVersion: CatalogVersionState;
}
function CatalogVersionCard(props: CatalogVersionCardProps) {
    const { catalogversion, refCatalogVersion } = props;
    return (
        <Card>
            <CardHeader className="flex gap-3">
                {catalogversion.spec.catalogType === 'config' && (
                    <FcSettings className="text-[#F6B519] text-3xl"/>
                )}
                {catalogversion.spec.catalogType === 'solutionversion' && (
                    <FcTemplate className="text-[#F6B519] text-3xl"/>
                )}
                <div className="flex flex-col">
                <p className="text-md">{catalogversion.spec.rootResource}:{catalogversion.metadata.name.split("-v-").pop()}</p>
                <p className="text-small text-default-500">{catalogversion.spec.catalogType}</p>
                </div>
            </CardHeader>
            <Divider/>
            <CardBody>
                {(catalogversion.spec.catalogType === 'config' && !catalogversion.spec.objectRef?.name) && (
                    <PropertyTable properties={catalogversion.spec.properties} refProperties={refCatalogVersion?.spec.properties} />
                )}
                {(catalogversion.spec.catalogType === 'config' && catalogversion.spec.objectRef?.name) && (
                    <div style={{ whiteSpace: 'nowrap' , display: 'inline-flex', gap: '0.5rem', color: 'darkolivegreen'}}><FaGithub />{catalogversion.spec.objectRef.address}</div>                    
                )}
                {catalogversion.spec.catalogType === 'solutionversion' && (
                    <SolutionVersionSpecCard solutionversion={catalogversion.spec.properties['spec']} />
                )}
            </CardBody>
            <Divider/>
            {catalogversion.spec.parentName != "" && catalogversion.spec.parentName != undefined && (
            
            <CardFooter>
                <span className='text-sm'>
                    <div>{`overrides: ${catalogversion.spec.parentName}`}</div></span>
            </CardFooter>
            )}
        </Card>
    );
}
export default CatalogVersionCard;