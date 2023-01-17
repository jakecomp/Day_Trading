import styled from '@emotion/styled'
import { Link } from 'react-router-dom'

export const SignInHeader = styled.div`
    font-size: 36px;
    color: #000000;
    line-height: 43.88px;
`

export const SignInText = styled.div`
    font-size: 16px;
    color: #7c80a5;
    line-height: 19.5px;
`

export const SignInLink = styled(Link)`
    font-size: 16px;
    color: #000000;
    line-height: 19.5px;
    &:hover {
        color: #4555e5;
    }
`

export const SignInputPlaceholder = styled.div`
    font-size: 14px;
    color: #7c80a5;
    line-height: 17.07px;
`

export const SignInputLabel = styled.label`
    font-size: 16px;
    color: #000000;
    line-height: 19.5px;
`
