'use client';

import {Card, CardHeader, CardBody, CardFooter, Divider, Link, Image} from '@nextui-org/react';
import {Chip} from "@nextui-org/react";
import { Site } from '../../app/types';
import {CheckIcon} from "../CheckIcon";
import { TbBuildingCommunity } from 'react-icons/tb';
interface SiteCardProps {
    site: Site;
}
function SiteCard(props: SiteCardProps) {
    const { site } = props;
    let cardClassName = '';
    if (site.self) {
        cardClassName = 'bg-primary-50';
    } else if (site.lastReported) {
        const lastReported = new Date(site.lastReported);
        const now = new Date();
        const diffInSeconds = (now.getTime() - lastReported.getTime()) / 1000;
        if (diffInSeconds <= 30) {
            cardClassName = 'bg-success-50';
        } else {
            cardClassName = 'bg-danger-50';
        }
    } else {
        cardClassName = 'bg-danger-50';
    }
    return (
        <Card className={`max-w-[400px] ${cardClassName}`}>
            <CardHeader className="flex gap-3 items-start">
                <span className='sitecard_icon'><TbBuildingCommunity /></span>
                <div className="flex flex-col">
                <p className="text-md">{site.name}</p>
                <p className="text-small text-default-500">{`${site.address} ${site.city}, ${site.state} ${site.zip} ${site.country}`}</p>
                </div>
            </CardHeader>
            <Divider/>
            <CardBody>
                
            </CardBody>
            <Divider/>
            <CardFooter className="justify-between">
                <Chip startContent={<CheckIcon size={18} />} color="primary" variant="light">{`Version: ${site.version}`}</Chip>
                {cardClassName === 'bg-success-50' && (
                    <Chip color="success">Online</Chip>
                )}
                {cardClassName === 'bg-danger-50' && (
                    <Chip color="danger">Offline</Chip>
                )}
                {cardClassName === 'bg-primary-50' && (
                    <Chip color="primary">Primay</Chip>
                )}              
            </CardFooter>
        </Card>
    );
}
export default SiteCard;