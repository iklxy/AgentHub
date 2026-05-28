/*这是一个计算器类 实现二元数的加减乘除开方乘方*/
#include<iostream>
#include<map>
#include<string>
#include<memory>
#include<stdexcept>
#include<cmath>
using namespace std;
//抽象类
class operation
{
    public:
    virtual ~operation()=default;//虚析构
    //获取结果
    virtual double getResult(double numberA,double numberB)=0;
    //获取运算符
    virtual string getOperator()=0;
};
//加法类
class AddOperation:public operation
{
    public:
    double getResult(double numberA,double numberB) override
    {
        return numberA+numberB;
    }

    string getOperator() override
    {
        return "+";
    }
};
//减法类
class SubOperation:public operation
{
    public:
    double getResult(double numberA,double numberB) override
    {
        return numberA-numberB;
    }

    string getOperator() override
    {
        return "-";
    }
};
//乘法类
class MulOperation:public operation
{
    public:
    double getResult(double numberA,double numberB) override
    {
        return numberA*numberB;
    }

    string getOperator() override
    {
        return "*";
    }
};
//除法类
class DivOperation:public operation
{
    public:
    double getResult(double numberA,double numberB) override
    {
        if(numberB==0)
        {
            throw runtime_error("the divisor cannot be zero");
        }
        return numberA/numberB;
    }

    string getOperator() override
    {
        return "/";
    }
};
//乘方类
class PowOperation:public operation
{
    public:
    double getResult(double numberA,double numberB) override
    {
        return pow(numberA,numberB);
    }
    string getOperator() override
    {
        return "^";
    }
};
//开方类
class SqrtOperation:public operation
{
    public:
    double getResult(double numberA,double numberB) override
    {
        if(numberA < 0)
        {
            throw runtime_error("the base cannot be negative");
        }
        return pow(numberA,1.0/numberB);
    }
    string getOperator() override
    {
        return "√";
    }
};
//计算器类
class calculator
{
    private:
    map<string,shared_ptr<operation> >operMap;
    public:
    calculator()
    {
        operMap["+"]=make_shared<AddOperation>();
        operMap["-"]=make_shared<SubOperation>();
        operMap["*"]=make_shared<MulOperation>();
        operMap["/"]=make_shared<DivOperation>();
        operMap["^"]=make_shared<PowOperation>();
        operMap["√"]=make_shared<SqrtOperation>();
    }

    double calculate(double numberA,double numberB,string oper) const
    {
        auto it=operMap.find(oper);
        if(it==operMap.end())
        {
            throw runtime_error("No such operator");
        }
        else
        {
            return it->second->getResult(numberA,numberB);
        }
    }
    void showOperators() const
    {
        cout<<"supported operators: ";
        for(const auto &pair:operMap)
        {
            cout<<pair.first<<" ";
        }
        cout<<endl;
    }
};
int main()
{
    calculator calc;
    calc.showOperators();
    double numA,numB;
    string oper;
    cout<<"please input first number: ";
    cin>>numA;
    cout<<"please input operator: ";
    cin>>oper;
    cout<<"please input second number: ";
    cin>>numB;
    try
    {
        double result=calc.calculate(numA,numB,oper);
        cout<<"resule"<<result<<endl;
    }
    catch(const exception &error)
    {
        cerr<<"error"<<error.what()<<endl;
    }
    system("pause");
    return 0;
}