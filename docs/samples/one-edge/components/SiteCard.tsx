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
                <p className="text-small text-default-500">{site.description}</p>
                </div>
            </CardHeader>
            <Divider/>
            <CardBody>
                <Table removeWrapper>
                    <TableHeader>
                        <TableColumn>SOLUTION</TableColumn>
                        <TableColumn>INSTANCE</TableColumn>
                        <TableColumn>STATUS</TableColumn>
                    </TableHeader>
                    <TableBody>
                        <TableRow key="1">
                            <TableCell>site-app</TableCell>
                            <TableCell>site-instance</TableCell>
                            <TableCell>
                                <Chip className="capitalize" color="success" size="sm" variant="flat">OK</Chip>
                            </TableCell>
                        </TableRow>
                        <TableRow key="2">
                            <TableCell>line-app</TableCell>
                            <TableCell>line-1</TableCell>
                            <TableCell>
                                <Chip className="capitalize" color="success" size="sm" variant="flat">OK</Chip>
                            </TableCell>
                        </TableRow>
                        <TableRow key="3">
                            <TableCell>line-app</TableCell>
                            <TableCell>line-2</TableCell>
                            <TableCell>
                                <Chip className="capitalize" color="success" size="sm" variant="flat">OK</Chip>
                            </TableCell>
                        </TableRow>                        
                    </TableBody>
                </Table>
            </CardBody>
            <Divider/>
            <CardFooter>
                <Link href={`tel:${site.phone}`}>Call: {site.phone}</Link>
            </CardFooter>
        </Card>
    );
}
export default SiteCard;