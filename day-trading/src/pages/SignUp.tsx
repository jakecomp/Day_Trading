import styled from '@emotion/styled'
import { BigBlackButton } from '../components/atoms/button'
import { SignBackground } from '../components/sign_in/background'
import { Card } from '../components/sign_in/card'
import { SignInHeader, SignInLink, SignInText } from '../components/sign_in/text'
import { HeaderContainer, LinkContainer } from '../components/sign_in/containers'
import { FieldForm, InputLabel, SignField } from '../components/sign_in/field'

export default function SignUp() {
    return (
        <div>
            <SignBackground>
                <Card>
                    <HeaderContainer>
                        <SignInHeader>Welcome to Day-trades</SignInHeader>
                        <LinkContainer>
                            <SignInText>Already have an account?</SignInText>
                            <SignInLink to='/signin'>Sign in</SignInLink>
                        </LinkContainer>
                    </HeaderContainer>

                    <FieldForm>
                        <InputLabel>First name</InputLabel>
                        <SignField placeholder='John'></SignField>
                        <InputLabel>Last name</InputLabel>
                        <SignField placeholder='Smith'></SignField>
                        <InputLabel>Email address</InputLabel>
                        <SignField placeholder='john.smith@email.com'></SignField>
                        <InputLabel>Password</InputLabel>
                        <SignField type='password' placeholder='Must be at least 6 characters'></SignField>
                    </FieldForm>

                    <BigBlackButton>Create account</BigBlackButton>
                </Card>
            </SignBackground>
        </div>
    )
}
