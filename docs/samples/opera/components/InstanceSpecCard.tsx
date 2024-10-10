import { InstanceSpec } from '../app/types';
import {Table, TableHeader, TableColumn, TableBody, TableRow, TableCell, Chip} from "@nextui-org/react";
import {FaDocker} from 'react-icons/fa';
import {SiHelm} from 'react-icons/si';
import {SiKubernetes} from 'react-icons/si';    
import {SiWindows} from 'react-icons/si';
interface InstanceSpecCardProps {
    instance: InstanceSpec;
}
function InstanceSpecCard(props: InstanceSpecCardProps) {
    const { instance } = props;
    return (
        <>
            <div className="components-header">Solution:</div>
            <div className="components-label">{instance.solution}</div>                        
            <Table removeWrapper>
                <TableHeader>
                    <TableColumn>TARGET</TableColumn>                    
                </TableHeader>
                <TableBody>
                    <TableRow>
                        <TableCell>
                            {instance.target.name}
                        </TableCell>
                    </TableRow>
                </TableBody>
            </Table>
        </>
    );
}
export default InstanceSpecCard;