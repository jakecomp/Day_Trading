import { BigBlackButton } from '../components/atoms/button'
import { SignBackground } from '../components/sign_in/background'
import { Card } from '../components/sign_in/card'
import { SignInHeader, SignInLink, SignInText } from '../components/sign_in/text'
import { InputContainer, LinkContainer } from '../components/sign_in/containers'
import { FieldForm, InputLabel, SignField } from '../components/sign_in/field'
import { useForm } from 'react-hook-form'

export const SignIn = () => {
    const { register, handleSubmit } = useForm()
    const RetrieveData = (data: any) => console.log(data)

    return (
        <div>
            <SignBackground>
                <Card>
                    <SignInHeader>Sign In to Day-trades</SignInHeader>
                    <FieldForm onSubmit={handleSubmit(RetrieveData)}>
                        <InputContainer>
                            <InputLabel>Username</InputLabel>
                            <SignField {...register('username')}></SignField>
                            <InputLabel>Password</InputLabel>
                            <SignField type='password' {...register('password')}></SignField>{' '}
                        </InputContainer>
                        <BigBlackButton>Sign in</BigBlackButton>
                    </FieldForm>
                    <LinkContainer>
                        <SignInText>New to Day-trades?</SignInText>
                        <SignInLink to='/'>Create an account</SignInLink>
                    </LinkContainer>
                </Card>
            </SignBackground>
        </div>
    )
}
