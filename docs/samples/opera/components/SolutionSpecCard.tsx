import { SolutionSpec } from '../app/types';
import {Table, TableHeader, TableColumn, TableBody, TableRow, TableCell, Chip} from "@nextui-org/react";
import {FaDocker} from 'react-icons/fa';
import {SiHelm} from 'react-icons/si';
import {SiKubernetes} from 'react-icons/si';    
import {SiWindows} from 'react-icons/si';
import { LuBinary } from 'react-icons/lu';
interface SolutionSpecCardProps {
    solution: SolutionSpec;
}
function SolutionSpecCard(props: SolutionSpecCardProps) {
    const { solution } = props;
    return (
        <>
            <div className="components-label">Components</div>
            <Table removeWrapper>
                <TableHeader>
                    <TableColumn> </TableColumn>
                    <TableColumn>NAME</TableColumn>   
                    <TableColumn>PACKAGE</TableColumn>      
                    <TableColumn>VERSION</TableColumn>        
                </TableHeader>
                <TableBody>
                    {solution.components.map((component: any) => (
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
                                {component.type === 'uwp' && (
                                    <SiWindows className="text-[#1111F9] text-xl"/>
                                )}
                                {component.properties['workload.type'] === 'binary' && (
                                    <LuBinary className="text-[#1111F9] text-xl"/>
                                )}
                            </TableCell>
                            <TableCell style={{ whiteSpace: 'nowrap' }}>{component.name}</TableCell>
                            <TableCell>
                                    {component.type === 'container' && component.properties?.['container.image'] &&  (
                                        <span style={{ whiteSpace: 'nowrap' }}>{component.properties['container.image'].split(':')[0]}</span>
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
                                    {component.properties['workload.name'] != undefined && (
                                        <span style={{ whiteSpace: 'nowrap' }}>{component.properties['workload.name']}</span>
                                    )}
                            </TableCell>
                            <TableCell>
                                    {component.type === 'container' && component.properties?.['container.image'] &&  (
                                        <span>
                                            {component.properties['container.image'].includes(':')
                                            ? component.properties['container.image'].split(':')[1]
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
                                    {component.properties['workload.type'] != undefined && (
                                        <span style={{ whiteSpace: 'nowrap' }}>{component.properties['workload.type']}</span>
                                    )}
                            </TableCell>
                        </TableRow>
                    ))}                            
                </TableBody>
            </Table>
        </>
    );
}
export default SolutionSpecCard;