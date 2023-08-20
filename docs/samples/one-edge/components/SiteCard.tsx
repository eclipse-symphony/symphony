'use client';

import {Card, CardHeader, CardBody, CardFooter, Divider, Link, Image} from '@nextui-org/react';
import {Table, TableHeader, TableColumn, TableBody, TableRow, TableCell, Chip} from "@nextui-org/react";

import { Site } from '../types';
interface SiteCardProps {
    site: Site;
}
function SiteCard(props: SiteCardProps) {
    const { site } = props;
    return (
        <Card className="max-w-[400px]">
            <CardHeader className="flex gap-3">
                <Image
                alt="nextui logo"
                height={40}
                radius="sm"
                src="https://avatars.githubusercontent.com/u/86160567?s=200&v=4"
                width={40}
                />
                <div className="flex flex-col">
                <p className="text-md">{site.name}</p>
                <p className="text-small text-default-500">{`${site.address} ${site.city}, ${site.state} ${site.zip} ${site.country}`}</p>
                </div>
            </CardHeader>
            <Divider/>
            <CardBody>
                
            </CardBody>
            <Divider/>
            <CardFooter>
                <p>{`Symphonyh Version: ${site.version}`}</p>
            </CardFooter>
        </Card>
    );
}
export default SiteCard;