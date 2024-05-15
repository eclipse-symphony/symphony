import { TargetSpec } from '../app/types';
import {Table, TableHeader, TableColumn, TableBody, TableRow, TableCell, Chip} from "@nextui-org/react";
import {FaDocker} from 'react-icons/fa';
import {SiHelm} from 'react-icons/si';
import {SiKubernetes} from 'react-icons/si';    
import {SiWindows} from 'react-icons/si';
import { SiGnubash } from "react-icons/si";
import { CgListTree } from "react-icons/cg";
interface TargetSpecCardProps {
    target: TargetSpec;
}
function TargetSpecCard(props: TargetSpecCardProps) {
    const { target } = props;
    return (
        <Table removeWrapper>
            <TableHeader>
                <TableColumn> </TableColumn>
                <TableColumn>NAME</TableColumn>   
                <TableColumn>PACKAGE</TableColumn>      
                <TableColumn>VERSION</TableColumn>        
            </TableHeader>
            <TableBody>
                {target.components && (target.components.map((component: any) => (
                    <TableRow key={component.name}>
                        <TableCell>
                            {component.type === 'container' && (
                                <FaDocker className="text-[#AAAAF9] text-xl"/>
                            )}
                            {component.type === 'helm.v3' && (
                                <SiHelm className="text-[#AAAAF9] text-xl"/>
                            )}
                            {component.type === 'yaml.k8s' && (
                                <SiKubernetes className="text-[#1111F9] text-xl"/>
                            )}
                            {component.type === 'program' && (
                                <CgListTree className="text-[#1111F9] text-xl"/>
                            )}
                            {component.type === 'uwp' && (
                                <SiWindows className="text-[#1111F9] text-xl"/>
                            )}
                             {component.type === 'script' && (
                                <SiGnubash className="text-[#1111F9] text-xl"/>
                            )}
                        </TableCell>
                        <TableCell style={{ whiteSpace: 'nowrap' }}>{component.name}</TableCell>
                        <TableCell>
                                {component.type === 'container' && (
                                    <span style={{ whiteSpace: 'nowrap' }}>{component.properties['container.image'].split(':')[0]}</span>
                                )}
                                {component.type === 'script' && (
                                    <span style={{ whiteSpace: 'nowrap' }}>{component.properties['container.image'].split(':')[0]}</span>
                                )}
                                {component.type === 'program' && (
                                    <span style={{ whiteSpace: 'nowrap' }}>{component.properties['program.image'].split(':')[0]}</span>
                                )}
                                {component.type === 'helm.v3' && (
                                    <span style={{ whiteSpace: 'nowrap' }}>{component.properties['chart']['repo']}</span>
                                )}
                                {component.type === 'yaml.k8s' && (
                                    <span style={{ whiteSpace: 'nowrap' }}>{`[object]`}</span>
                                )}
                                {component.type === 'uwp' && (
                                    <span style={{ whiteSpace: 'nowrap' }}>{component.properties['app.image']}</span>
                                )}
                        </TableCell>
                        <TableCell>
                                {component.type === 'container' && (
                                    <span>
                                        {component.properties['container.image'].includes(':')
                                        ? component.properties['container.image'].split(':')[1]
                                        : '(latest)'}
                                    </span>
                                )}
                                {component.type === 'script' && (
                                    <span>
                                        {component.properties['container.image'].includes(':')
                                        ? component.properties['container.image'].split(':')[1]
                                        : '(latest)'}
                                    </span>
                                )}
                                {component.type === 'program' && (
                                    <span>
                                        {component.properties['program.image'].includes(':')
                                        ? component.properties['program.image'].split(':')[1]
                                        : '(latest)'}
                                    </span>
                                )}
                                {component.type === 'helm.v3' && (
                                    <span>{component.properties['chart']['version']}</span>
                                )}
                                {component.type === 'yaml.k8s' && (
                                    <span>{`n/a`}</span>
                                )}
                                {component.type === 'uwp' && (
                                    <span>{component.properties['app.version']}</span>
                                )}
                        </TableCell>
                    </TableRow>
                )))}                            
            </TableBody>
        </Table>
    );
}
export default TargetSpecCard;