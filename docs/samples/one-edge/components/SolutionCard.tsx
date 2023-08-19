import { Solution } from '../types';
import {Table, TableHeader, TableColumn, TableBody, TableRow, TableCell, Chip} from "@nextui-org/react";
import {FaDocker} from 'react-icons/fa';

interface SolutionCardProps {
    solution: Solution;
}
function SolutionCard(props: SolutionCardProps) {
    const { solution } = props;
    return (
        <Table removeWrapper>
            <TableHeader>
                <TableColumn> </TableColumn>
                <TableColumn>NAME</TableColumn>   
                <TableColumn>IMAGE</TableColumn>              
            </TableHeader>
            <TableBody>
                {solution.spec.components.map((component: any) => (
                    <TableRow key={component.name}>
                        <TableCell>
                            {component.type === 'container' && (
                                <FaDocker className="text-[#AAAAF9] text-xl"/>
                            )}
                        </TableCell>
                        <TableCell>{component.name}</TableCell>
                        <TableCell>{component.properties['container.image']}</TableCell>
                    </TableRow>
                ))}                            
            </TableBody>
        </Table>
    );
}
export default SolutionCard;