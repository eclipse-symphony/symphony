'use client';

import {Card, CardHeader, CardBody, CardFooter, Divider, Link, Image} from '@nextui-org/react';
import { CatalogVersionState } from '../../app/types';
import PropertyTable from '../PropertyTable';
import SolutionVersionSpecCard from '../SolutionVersionSpecCard';
import { FcSettings, FcTemplate } from 'react-icons/fc';
import { FaGithub } from 'react-icons/fa';

interface AssetCardProps {
    catalogversion: CatalogVersionState;
    refCatalogVersion?: CatalogVersionState | null | undefined;
}
function AssetCard(props: AssetCardProps) {
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
                <p className="text-md">{catalogversion.spec.name}</p>
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
            {catalogversion.spec.metadata?.['override'] && (
            
            <CardFooter>
                <span className='text-sm'>
                    <div>{`overrides: ${catalogversion.spec.metadata['override']}`}</div></span>
            </CardFooter>
            )}
        </Card>
    );
}
export default AssetCard;